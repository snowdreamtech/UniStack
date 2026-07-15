// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiHandler(t *testing.T) {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	h1 := slog.NewJSONHandler(&buf1, &slog.HandlerOptions{Level: slog.LevelInfo})
	h2 := slog.NewTextHandler(&buf2, &slog.HandlerOptions{Level: slog.LevelError})

	multi := NewMultiHandler(h1, h2)
	logger := slog.New(multi)

	// Test 1: Info level message
	// Should go to buf1, but NOT buf2 (since h2 is LevelError)
	logger.Info("info message")

	assert.Contains(t, buf1.String(), "info message")
	assert.Empty(t, buf2.String(), "buf2 should be empty because h2 is Error level")

	buf1.Reset()
	buf2.Reset()

	// Test 2: Error level message
	// Should go to both
	logger.Error("error message")

	assert.Contains(t, buf1.String(), "error message")
	assert.Contains(t, buf2.String(), "error message")

	buf1.Reset()
	buf2.Reset()

	// Test 3: WithAttrs
	l2 := logger.With("key", "value")
	l2.Error("msg with attr")

	assert.Contains(t, buf1.String(), "value")
	assert.Contains(t, buf2.String(), "value")

	var jsonMap map[string]interface{}
	err := json.Unmarshal(buf1.Bytes(), &jsonMap)
	assert.NoError(t, err)
	assert.Equal(t, "value", jsonMap["key"])

	buf1.Reset()
	buf2.Reset()

	// Test 4: WithGroup
	l3 := logger.WithGroup("mygroup").With("gkey", "gval")
	l3.Error("grouped message")

	err = json.Unmarshal(buf1.Bytes(), &jsonMap)
	assert.NoError(t, err)

	group, ok := jsonMap["mygroup"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "gval", group["gkey"])
}

func TestMultiHandler_Enabled(t *testing.T) {
	// Not passing nil writer as that panics in slog occasionally, use a buffer
	var b bytes.Buffer
	h1 := slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelError})
	h2 := slog.NewTextHandler(&b, &slog.HandlerOptions{Level: slog.LevelWarn})

	multi := NewMultiHandler(h1, h2)
	ctx := context.Background()

	// Neither supports Debug or Info
	assert.False(t, multi.Enabled(ctx, slog.LevelDebug))
	assert.False(t, multi.Enabled(ctx, slog.LevelInfo))

	// h2 supports Warn
	assert.True(t, multi.Enabled(ctx, slog.LevelWarn))

	// Both support Error
	assert.True(t, multi.Enabled(ctx, slog.LevelError))
}
