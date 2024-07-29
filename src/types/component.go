// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package types contains all the types used by Zarf.
package types

import (
	"github.com/invopop/jsonschema"
	"github.com/zarf-dev/zarf/src/pkg/utils/exec"
	"github.com/zarf-dev/zarf/src/pkg/variables"
	"github.com/zarf-dev/zarf/src/types/extensions"
)

// ZarfComponent is the primary functional grouping of assets to deploy by Zarf.
type ZarfComponent struct {
	// The name of the component
	Name string `jsonschema:"required,pattern=^[a-z0-9][a-z0-9\\-]*$"`

	// Message to include during package deploy describing the purpose of this component
	Description string

	// Determines the default Y/N state for installing this component on package deploy
	Default bool

	// Do not prompt user to install this component
	Required *bool

	// Filter when this component is included in package creation or deployment
	Only ZarfComponentOnlyTarget

	// [Deprecated] Create a user selector field based on all components in the same group. This will be removed in Zarf v1.0.0. Consider using 'only.flavor' instead.
	DeprecatedGroup string `json:"group" jsonschema:"deprecated=true"`

	// [Deprecated] Specify a path to a public key to validate signed online resources. This will be removed in Zarf v1.0.0.
	DeprecatedCosignKeyPath string `json:"cosignKeyPath" jsonschema:"deprecated=true"`

	// Import a component from another Zarf package
	Import ZarfComponentImport

	// Kubernetes manifests to be included in a generated Helm chart on package deploy
	Manifests []ZarfManifest

	// Helm charts to install during package deploy
	Charts []ZarfChart

	// Datasets to inject into a container in the target cluster
	DataInjections []ZarfDataInjection

	// Files or folders to place on disk during package deployment
	Files []ZarfFile

	// List of OCI images to include in the package
	Images []string

	// List of git repos to include in the package
	Repos []string

	// Extend component functionality with additional features
	Extensions extensions.ZarfComponentExtensions

	// [Deprecated] (replaced by actions) Custom commands to run before or after package deployment. This will be removed in Zarf v1.0.0.
	DeprecatedScripts DeprecatedZarfComponentScripts `json:"scripts" jsonschema:"deprecated=true"`

	// Custom commands to run at various stages of a package lifecycle
	Actions ZarfComponentActions
}

// RequiresCluster returns if the component requires a cluster connection to deploy
func (c ZarfComponent) RequiresCluster() bool {
	hasImages := len(c.Images) > 0
	hasCharts := len(c.Charts) > 0
	hasManifests := len(c.Manifests) > 0
	hasRepos := len(c.Repos) > 0
	hasDataInjections := len(c.DataInjections) > 0

	if hasImages || hasCharts || hasManifests || hasRepos || hasDataInjections {
		return true
	}

	return false
}

// IsRequired returns if the component is required or not.
func (c ZarfComponent) IsRequired() bool {
	if c.Required != nil {
		return *c.Required
	}

	return false
}

// ZarfComponentOnlyTarget filters a component to only show it for a given local OS and cluster.
type ZarfComponentOnlyTarget struct {
	// Only deploy component to specified OS
	LocalOS string `jsonschema:"enum=linux,enum=darwin,enum=windows"`
	// Only deploy component to specified clusters
	Cluster ZarfComponentOnlyCluster
	// Only include this component when a matching '--flavor' is specified on 'zarf package create'
	Flavor string
}

// JSONSchemaExtend extends the generated json schema during `zarf internal gen-config-schema`
func (ZarfComponentOnlyTarget) JSONSchemaExtend(schema *jsonschema.Schema) {
	// Need this since it auto camel cases the acronym
	localOS := schema.Properties.GetPair("localOs")
	localOS.Key = "localOS"
	schema.Properties.Delete("localOs")
	schema.Properties.AddPairs(*localOS)
}

// ZarfComponentOnlyCluster represents the architecture and K8s cluster distribution to filter on.
type ZarfComponentOnlyCluster struct {
	// Only create and deploy to clusters of the given architecture
	Architecture string `jsonschema:"enum=amd64,enum=arm64"`
	// A list of kubernetes distros this package works with (Reserved for future use)
	Distros []string `jsonschema:"example=k3s,example=eks"`
}

// ZarfFile defines a file to deploy.
type ZarfFile struct {
	// Local folder or file path or remote URL to pull into the package
	Source string `jsonschema:"required"`
	// (files only) Optional SHA256 checksum of the file
	Shasum string
	// The absolute or relative path where the file or folder should be copied to during package deploy
	Target string `jsonschema:"required"`
	// (files only) Determines if the file should be made executable during package deploy
	Executable bool
	// List of symlinks to create during package deploy
	Symlinks []string
	// Local folder or file to be extracted from a 'source' archive
	ExtractPath string
}

// ZarfChart defines a helm chart to be deployed.
type ZarfChart struct {
	// The name of the chart within Zarf; note that this must be unique and does not need to be the same as the name in the chart repo
	Name string `jsonschema:"required"`
	// The version of the chart to deploy; for git-based charts this is also the tag of the git repo by default (when not using the '@' syntax for 'repos')
	Version string
	// The URL of the OCI registry, chart repository, or git repo where the helm chart is stored
	URL string `jsonschema:"example=OCI registry: oci://ghcr.io/stefanprodan/charts/podinfo,example=helm chart repo: https://stefanprodan.github.io/podinfo,example=git repo: https://github.com/stefanprodan/podinfo (note the '@' syntax for 'repos' is supported here too)"`
	// The name of a chart within a Helm repository (defaults to the Zarf name of the chart)
	RepoName string
	// (git repo only) The sub directory to the chart within a git repo
	GitPath string `jsonschema:"example=charts/your-chart"`
	// The path to a local chart's folder or .tgz archive
	LocalPath string
	// The namespace to deploy the chart to
	Namespace string
	// The name of the Helm release to create (defaults to the Zarf name of the chart)
	ReleaseName string
	// Whether to not wait for chart resources to be ready before continuing
	NoWait bool
	// List of local values file paths or remote URLs to include in the package; these will be merged together when deployed
	ValuesFiles []string
	// [alpha] List of variables to set in the Helm chart
	Variables []ZarfChartVariable
}

// ZarfChartVariable represents a variable that can be set for a Helm chart overrides.
type ZarfChartVariable struct {
	// The name of the variable
	Name string `jsonschema:"required,pattern=^[A-Z0-9_]+$"`
	// A brief description of what the variable controls
	Description string `jsonschema:"required"`
	// The path within the Helm chart values where this variable applies
	Path string `jsonschema:"required"`
}

// ZarfManifest defines raw manifests Zarf will deploy as a helm chart.
type ZarfManifest struct {
	// A name to give this collection of manifests; this will become the name of the dynamically-created helm chart
	Name string `jsonschema:"required"`
	// The namespace to deploy the manifests to
	Namespace string
	// List of local K8s YAML files or remote URLs to deploy (in order)
	Files []string
	// Allow traversing directory above the current directory if needed for kustomization
	KustomizeAllowAnyDirectory bool
	// List of local kustomization paths or remote URLs to include in the package
	Kustomizations []string
	// Whether to not wait for manifest resources to be ready before continuing
	NoWait bool
}

// DeprecatedZarfComponentScripts are scripts that run before or after a component is deployed
type DeprecatedZarfComponentScripts struct {
	// Show the output of the script during package deployment
	ShowOutput bool
	// Timeout in seconds for the script
	TimeoutSeconds int
	// Retry the script if it fails
	Retry bool
	// Scripts to run before the component is added during package create
	Prepare []string
	// Scripts to run before the component is deployed
	Before []string
	// Scripts to run after the component successfully deploys
	After []string
}

// ZarfComponentActions are ActionSets that map to different zarf package operations
type ZarfComponentActions struct {
	// Actions to run during package creation
	OnCreate ZarfComponentActionSet
	// Actions to run during package deployment
	OnDeploy ZarfComponentActionSet
	// Actions to run during package removal
	OnRemove ZarfComponentActionSet
}

// ZarfComponentActionSet is a set of actions to run during a zarf package operation
type ZarfComponentActionSet struct {
	// Default configuration for all actions in this set
	Defaults ZarfComponentActionDefaults
	// Actions to run at the start of an operation
	Before []ZarfComponentAction
	// Actions to run at the end of an operation
	After []ZarfComponentAction
	// Actions to run if all operations succeed
	OnSuccess []ZarfComponentAction
	// Actions to run if all operations fail
	OnFailure []ZarfComponentAction
}

// ZarfComponentActionDefaults sets the default configs for child actions
type ZarfComponentActionDefaults struct {
	// Hide the output of commands during execution (default false)
	Mute bool
	// Default timeout in seconds for commands (default to 0, no timeout)
	MaxTotalSeconds int
	// Retry commands given number of times if they fail (default 0)
	MaxRetries int
	// Working directory for commands (default CWD)
	Dir string
	// Additional environment variables for commands
	Env []string
	// (cmd only) Indicates a preference for a shell for the provided cmd to be executed in on supported operating systems
	Shell exec.Shell
}

// ZarfComponentAction represents a single action to run during a zarf package operation
type ZarfComponentAction struct {
	// Hide the output of the command during package deployment (default false)
	Mute *bool
	// Timeout in seconds for the command (default to 0, no timeout for cmd actions and 300, 5 minutes for wait actions)
	MaxTotalSeconds *int
	// Retry the command if it fails up to given number of times (default 0)
	MaxRetries *int
	// The working directory to run the command in (default is CWD)
	Dir *string
	// Additional environment variables to set for the command
	Env []string
	// The command to run. Must specify either cmd or wait for the action to do anything.
	Cmd string
	// (cmd only) Indicates a preference for a shell for the provided cmd to be executed in on supported operating systems
	Shell *exec.Shell
	// [Deprecated] (replaced by setVariables) (onDeploy/cmd only) The name of a variable to update with the output of the command. This variable will be available to all remaining actions and components in the package. This will be removed in Zarf v1.0.0
	DeprecatedSetVariable string `json:"setVariable" jsonschema:"pattern=^[A-Z0-9_]+$"`
	// (onDeploy/cmd only) An array of variables to update with the output of the command. These variables will be available to all remaining actions and components in the package.
	SetVariables []variables.Variable
	// Description of the action to be displayed during package execution instead of the command
	Description string
	// Wait for a condition to be met before continuing. Must specify either cmd or wait for the action. See the 'zarf tools wait-for' command for more info.
	Wait *ZarfComponentActionWait
}

// ZarfComponentActionWait specifies a condition to wait for before continuing
type ZarfComponentActionWait struct {
	// Wait for a condition to be met in the cluster before continuing. Only one of cluster or network can be specified.
	Cluster *ZarfComponentActionWaitCluster
	// Wait for a condition to be met on the network before continuing. Only one of cluster or network can be specified.
	Network *ZarfComponentActionWaitNetwork
}

// ZarfComponentActionWaitCluster specifies a condition to wait for before continuing
type ZarfComponentActionWaitCluster struct {
	// The kind of resource to wait for
	Kind string `jsonschema:"required,example=Pod,example=Deployment"`
	// The name of the resource or selector to wait for
	Name string `jsonschema:"required,example=podinfo,example=app=podinfo"`
	// The namespace of the resource to wait for
	Namespace string
	// The condition or jsonpath state to wait for; defaults to exist, a special condition that will wait for the resource to exist
	Condition string `jsonschema:"example=Ready,example=Available,'{.status.availableReplicas}'=23"`
}

// ZarfComponentActionWaitNetwork specifies a condition to wait for before continuing
type ZarfComponentActionWaitNetwork struct {
	// The protocol to wait for
	Protocol string `jsonschema:"required,enum=tcp,enum=http,enum=https"`
	// The address to wait for
	Address string `jsonschema:"required,example=localhost:8080,example=1.1.1.1"`
	// The HTTP status code to wait for if using http or https
	Code int `jsonschema:"required,example=200,example=404"`
}

// ZarfContainerTarget defines the destination info for a ZarfData target
type ZarfContainerTarget struct {
	// The namespace to target for data injection
	Namespace string `jsonschema:"required"`
	// The K8s selector to target for data injection
	Selector string `jsonschema:"required,example=app=data-injection"`
	// The container name to target for data injection
	Container string `jsonschema:"required"`
	// The path within the container to copy the data into
	Path string `jsonschema:"required"`
}

// ZarfDataInjection is a data-injection definition.
type ZarfDataInjection struct {
	// Either a path to a local folder/file or a remote URL of a file to inject into the given target pod + container
	Source string `jsonschema:"required"`
	// The target pod + container to inject the data into
	Target ZarfContainerTarget `jsonschema:"required"`
	// Compress the data before transmitting using gzip. Note: this requires support for tar/gzip locally and in the target image.
	Compress bool
}

// ZarfComponentImport structure for including imported Zarf components.
type ZarfComponentImport struct {
	// The name of the component to import from the referenced zarf.yaml
	Name string
	// The path to the directory containing the zarf.yaml to import
	Path string
	// [beta] The URL to a Zarf package to import via OCI
	URL string `jsonschema:"pattern=^oci://.*$"`
}

// JSONSchemaExtend extends the generated json schema during `zarf internal gen-config-schema`
func (ZarfComponentImport) JSONSchemaExtend(schema *jsonschema.Schema) {
	path, _ := schema.Properties.Get("path")
	url, _ := schema.Properties.Get("url")

	notSchema := &jsonschema.Schema{
		Pattern: ZarfPackageTemplatePrefix,
	}

	path.Not = notSchema
	url.Not = notSchema
}
