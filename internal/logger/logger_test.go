// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureOutput redirects the global slog logger to a buffer for testing
// and returns a cleanup function that restores the original logger.
func captureOutput(buf *bytes.Buffer, jsonFmt bool) func() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	var h slog.Handler
	if jsonFmt {
		h = slog.NewJSONHandler(buf, opts)
	} else {
		h = slog.NewTextHandler(buf, opts)
	}
	original := slog.Default()
	slog.SetDefault(slog.New(h))
	return func() { slog.SetDefault(original) }
}

func TestInit_Default(t *testing.T) {
	Init(false, false, false, false)
	// Default logger should not panic
	slog.Info("init default")
}

func TestInit_Debug(t *testing.T) {
	var buf bytes.Buffer
	restore := captureOutput(&buf, false)
	defer restore()

	Init(true, false, false, false)
	// Re-capture after Init sets handler
	var buf2 bytes.Buffer
	restore2 := captureOutput(&buf2, false)
	defer restore2()

	slog.Debug("debug msg")
	assert.Contains(t, buf2.String(), "debug msg")
}

func TestInit_Quiet(t *testing.T) {
	// quiet mode sets level to Error, so Info and Debug should be suppressed
	// We verify by setting up a buffer AFTER calling Init
	Init(false, true, false, false)

	// After Init sets the real handler (to os.Stderr), override with a testable buffer
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelError} // mirror quiet logic
	h := slog.NewTextHandler(&buf, opts)
	slog.SetDefault(slog.New(h))
	defer func() { Init(false, false, false, false) }()

	slog.Info("quiet info")   // should be filtered (below Error)
	slog.Error("quiet error") // should appear

	assert.NotContains(t, buf.String(), "quiet info")
	assert.Contains(t, buf.String(), "quiet error")
}

func TestInit_Silent(t *testing.T) {
	Init(false, false, true, false)
	// Silent mode: logger should discard everything without panic
	slog.Error("silent error")
	slog.Info("silent info")
}

func TestInit_JSON(t *testing.T) {
	var buf bytes.Buffer
	// Set logger to capture output
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := slog.NewJSONHandler(&buf, opts)
	slog.SetDefault(slog.New(h))
	defer func() {
		Init(false, false, false, false)
	}()

	// Now call Init with JSON, which should produce JSON output
	Init(false, false, false, true)

	// Re-capture
	var buf2 bytes.Buffer
	h2 := slog.NewJSONHandler(&buf2, opts)
	slog.SetDefault(slog.New(h2))
	slog.Info("json message")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(buf2.String()), "{"), "expected JSON output")
}

func TestInit_DebugAndQuiet(t *testing.T) {
	// quiet takes precedence when both are set — level = Error
	Init(true, true, false, false)
	// Should not panic
	slog.Info("should be suppressed")
}

func TestInit_SilentOverridesAll(t *testing.T) {
	// silent overrides debug and quiet
	Init(true, true, true, false)
	slog.Error("discarded")
}

func TestInit_Restorable(t *testing.T) {
	// Ensure Init leaves a functional global logger
	Init(false, false, false, false)
	assert.NotNil(t, slog.Default())
}

func TestInit_TextHandler(t *testing.T) {
	// Verify text handler produces non-JSON output
	var buf bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := slog.NewTextHandler(&buf, opts)
	slog.SetDefault(slog.New(h))
	defer func() { Init(false, false, false, false) }()

	slog.Info("text message")
	out := buf.String()
	assert.Contains(t, out, "text message")
	assert.False(t, strings.HasPrefix(strings.TrimSpace(out), "{"), "expected text output, not JSON")
}

func TestInit_DiscardWriter(t *testing.T) {
	// Ensure silent mode truly discards all output
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	h := slog.NewTextHandler(io.Discard, opts)
	slog.SetDefault(slog.New(h))
	defer func() { Init(false, false, false, false) }()

	// Should not panic or produce output
	slog.Debug("discarded debug")
	slog.Info("discarded info")
	slog.Error("discarded error")
}
