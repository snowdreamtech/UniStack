// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/snowdreamtech/unistack/internal/env"
	"gopkg.in/natefinch/lumberjack.v2"
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

	// 2. Determine output format (JSON vs Text/Pterm)
	var consoleHandler slog.Handler
	if jsonFmt {
		consoleHandler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		pterm.EnableDebugMessages()
		consoleHandler = NewPtermHandler(level)
	}

	// 3. Setup File Track (JSONL with Lumberjack Rotation)
	logDir := filepath.Join(env.GetDataDir(), "logs")
	os.MkdirAll(logDir, 0700)
	logFile := filepath.Join(logDir, "unistack.jsonl")

	fileWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	}

	fileHandler := slog.NewJSONHandler(fileWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Always capture everything in the file track
	})

	// 4. Combine into MultiHandler
	multiHandler := NewMultiHandler(consoleHandler, fileHandler)

	// 5. Set global logger
	logger := slog.New(multiHandler)
	slog.SetDefault(logger)
}
