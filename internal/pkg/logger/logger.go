// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"io"
	"log/slog"
	"os"
)

// Init configures the global slog default logger based on the provided flags.
// It uses zero external dependencies and maps neatly to CLI standard behavior.
func Init(debug, quiet, silent, jsonFmt bool) {
	// 0. If silent, discard all output
	if silent {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return
	}

	// 1. Determine log level
	var level slog.Level
	if quiet {
		level = slog.LevelError
	} else if debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	// 2. Determine output format (JSON vs Text)
	var handler slog.Handler
	if jsonFmt {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	// 3. Set global logger
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
