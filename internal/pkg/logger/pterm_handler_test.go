// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestPtermHandler(t *testing.T) {
	handler := NewPtermHandler(slog.LevelInfo)
	if handler == nil {
		t.Fatal("NewPtermHandler returned nil")
	}

	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Handler should be enabled for Info")
	}

	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Handler should not be enabled for Debug when level is Info")
	}

	// Test Handle without attrs
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := handler.Handle(context.Background(), rec)
	if err != nil {
		t.Errorf("Handle returned error: %v", err)
	}

	// Test Handle with attrs
	rec.AddAttrs(slog.String("key", "value"), slog.Int("num", 42))
	err = handler.Handle(context.Background(), rec)
	if err != nil {
		t.Errorf("Handle returned error: %v", err)
	}

	// Test levels
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, lvl := range levels {
		rec := slog.NewRecord(time.Now(), lvl, "test level", 0)
		_ = handler.Handle(context.Background(), rec)
	}

	// Test WithAttrs and WithGroup (they just return the receiver currently)
	h2 := handler.WithAttrs([]slog.Attr{slog.String("foo", "bar")})
	if h2 != handler {
		t.Error("WithAttrs should return the same handler")
	}

	h3 := handler.WithGroup("mygroup")
	if h3 != handler {
		t.Error("WithGroup should return the same handler")
	}
}
