// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/snowdreamtech/unigo/internal/database"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a temporary SQLite database for testing and returns the connection and a cleanup function.
func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "unigo_test_*")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")
	config := database.Config{
		Path:    dbPath,
		WALMode: true,
	}

	db, err := database.Open(context.Background(), config)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, cleanup
}
