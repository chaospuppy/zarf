// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package logging

import (
	"context"
	"log/slog"

	"github.com/zarf-dev/zarf/src/pkg/message"
)

type PtermHandler struct {
	attrs []slog.Attr
	group string
}

func NewPtermHandler() *PtermHandler {
	return &PtermHandler{}
}

func (h *PtermHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *PtermHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		message.Debug(r.Message)
	case slog.LevelInfo:
		message.Info(r.Message)
	case slog.LevelWarn:
		message.Warn(r.Message)
	case slog.LevelError:
		message.Warn(r.Message)
	}
	return nil
}

func (h *PtermHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PtermHandler{
		attrs: append(h.attrs, attrs...),
		group: h.group,
	}
}

func (h *PtermHandler) WithGroup(name string) slog.Handler {
	return &PtermHandler{
		attrs: h.attrs,
		group: name,
	}
}
