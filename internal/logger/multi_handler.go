// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"log/slog"
)

// MultiHandler is a slog.Handler that duplicates its log records to multiple underlying handlers.
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler creates a new MultiHandler with the given slog handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{
		handlers: handlers,
	}
}

// Enabled reports whether the handler handles records at the given level.
// It returns true if ANY of the underlying handlers returns true.
func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle calls Handle on all underlying handlers that are enabled for the record's level.
// It returns the first error encountered, if any.
func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if e := handler.Handle(ctx, r.Clone()); e != nil && err == nil {
				err = e // return the first error
			}
		}
	}
	return err
}

// WithAttrs returns a new MultiHandler whose underlying handlers include the given attributes.
func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var newHandlers []slog.Handler
	for _, handler := range h.handlers {
		newHandlers = append(newHandlers, handler.WithAttrs(attrs))
	}
	return NewMultiHandler(newHandlers...)
}

// WithGroup returns a new MultiHandler whose underlying handlers have the given group name.
func (h *MultiHandler) WithGroup(name string) slog.Handler {
	var newHandlers []slog.Handler
	for _, handler := range h.handlers {
		newHandlers = append(newHandlers, handler.WithGroup(name))
	}
	return NewMultiHandler(newHandlers...)
}
