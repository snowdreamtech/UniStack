// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"log/slog"

	"github.com/pterm/pterm"
)

// PtermHandler is a custom slog.Handler that routes log records to pterm.
type PtermHandler struct {
	level slog.Level
}

// NewPtermHandler creates a new PtermHandler with the specified log level.
func NewPtermHandler(level slog.Level) *PtermHandler {
	return &PtermHandler{
		level: level,
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *PtermHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle handles the Record by mapping slog levels to pterm equivalents.
func (h *PtermHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message

	// Format attributes if there are any
	if r.NumAttrs() > 0 {
		msg += " ["
		first := true
		r.Attrs(func(a slog.Attr) bool {
			if !first {
				msg += ", "
			}
			msg += a.Key + "=" + a.Value.String()
			first = false
			return true
		})
		msg += "]"
	}

	// Route to the appropriate pterm printer based on level
	switch {
	case r.Level >= slog.LevelError:
		pterm.Error.Println(msg)
	case r.Level >= slog.LevelWarn:
		pterm.Warning.Println(msg)
	case r.Level >= slog.LevelInfo:
		pterm.Info.Println(msg)
	default:
		pterm.Debug.Println(msg)
	}

	return nil
}

// WithAttrs returns a new Handler with additional attributes.
// For simplicity in CLI usage, we don't store persistent attrs, but we could if needed.
func (h *PtermHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns a new Handler with the given group appended to the receiver's existing groups.
func (h *PtermHandler) WithGroup(name string) slog.Handler {
	return h
}
