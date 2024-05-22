// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/internal/agent/http/admission"
	"github.com/defenseunicorns/zarf/src/internal/agent/operations"
	"github.com/defenseunicorns/zarf/src/types"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func createPodAdmissionRequest(t *testing.T, op v1.Operation, pod *corev1.Pod) *v1.AdmissionRequest {
	t.Helper()
	raw, err := json.Marshal(pod)
	require.NoError(t, err)
	return &v1.AdmissionRequest{
		Operation: op,
		Object: runtime.RawExtension{
			Raw: raw,
		},
	}
}

func TestPodMutationWebhook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	state := &types.ZarfState{RegistryInfo: types.RegistryInfo{Address: "127.0.0.1:31999"}}
	c := createTestClientWithZarfState(ctx, t, state)
	handler := admission.NewHandler().Serve(NewPodMutationHook(ctx, c))

	tests := []struct {
		name          string
		admissionReq  *v1.AdmissionRequest
		expectedPatch []operations.PatchOperation
		code          int
	}{
		{
			name: "pod with label should be mutated",
			admissionReq: createPodAdmissionRequest(t, v1.Create, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"should-be": "mutated"},
				},
				Spec: corev1.PodSpec{
					Containers:     []corev1.Container{{Image: "nginx"}},
					InitContainers: []corev1.Container{{Image: "busybox"}},
					EphemeralContainers: []corev1.EphemeralContainer{
						{
							EphemeralContainerCommon: corev1.EphemeralContainerCommon{
								Image: "alpine",
							},
						},
					},
				},
			}),
			expectedPatch: []operations.PatchOperation{
				operations.ReplacePatchOperation(
					"/spec/imagePullSecrets",
					[]corev1.LocalObjectReference{{Name: config.ZarfImagePullSecretName}},
				),
				operations.ReplacePatchOperation(
					"/spec/initContainers/0/image",
					"127.0.0.1:31999/library/busybox:latest-zarf-2140033595",
				),
				operations.ReplacePatchOperation(
					"/spec/ephemeralContainers/0/image",
					"127.0.0.1:31999/library/alpine:latest-zarf-1117969859",
				),
				operations.ReplacePatchOperation(
					"/spec/containers/0/image",
					"127.0.0.1:31999/library/nginx:latest-zarf-3793515731",
				),
				operations.ReplacePatchOperation(
					"/metadata/labels/zarf-agent",
					"patched",
				),
			},
			code: http.StatusOK,
		},
		{
			name: "pod with zarf-agent patched label should not be mutated",
			admissionReq: createPodAdmissionRequest(t, v1.Create, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"zarf-agent": "patched"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "nginx"}},
				},
			}),
			expectedPatch: nil,
			code:          http.StatusOK,
		},
		{
			name: "pod with no labels should not error",
			admissionReq: createPodAdmissionRequest(t, v1.Create, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Labels: nil,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Image: "nginx"}},
				},
			}),
			expectedPatch: []operations.PatchOperation{
				operations.ReplacePatchOperation(
					"/spec/imagePullSecrets",
					[]corev1.LocalObjectReference{{Name: config.ZarfImagePullSecretName}},
				),
				operations.ReplacePatchOperation(
					"/spec/containers/0/image",
					"127.0.0.1:31999/library/nginx:latest-zarf-3793515731",
				),
				operations.AddPatchOperation(
					"/metadata/labels",
					map[string]string{"zarf-agent": "patched"},
				),
			},
			code: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := sendAdmissionRequest(t, tt.admissionReq, handler, tt.code)
			if tt.expectedPatch == nil {
				require.Empty(t, string(resp.Patch))
			} else {
				expectedPatchJSON, err := json.Marshal(tt.expectedPatch)
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.True(t, resp.Allowed)
				require.JSONEq(t, string(expectedPatchJSON), string(resp.Patch))
			}
		})
	}
}