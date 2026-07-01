// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package transaction

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/snowdreamtech/unigo/internal/database"
	"github.com/snowdreamtech/unigo/internal/repository"
	"github.com/snowdreamtech/unigo/internal/repository/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "unigo_test.db")

	config := database.Config{
		Path:    dbPath,
		WALMode: true,
	}

	db, err := database.Open(context.Background(), config)
	require.NoError(t, err, "failed to open database")

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestNewSQLiteTransactionManager(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	assert.NotNil(t, tm)
	assert.IsType(t, &sqliteTransactionManager{}, tm)
}

func TestTransactionManager_Begin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())

	ctx := context.Background()
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)
	require.NotNil(t, tx)

	assert.NotNil(t, tx.CacheRepo())
	assert.NotNil(t, tx.AuditRepo())

	// Rollback to clean up
	err = tx.Rollback()
	assert.NoError(t, err)
}

func TestTransaction_Commit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Begin transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Add an audit entry
	entry := &repository.AuditEntry{
		Timestamp:       time.Now(),
		Operation:       "test_operation",
		Tool:            "test-tool",
		Version:         "1.0.0",
		Status:          "success",
		GpgVerification: "skipped",
	}

	err = tx.AuditRepo().Log(ctx, entry)
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify data was saved
	repo, _ := sqlite.NewAuditRepository(db.Conn())
	entries, err := repo.Query(ctx, repository.AuditFilter{})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "test_operation", entries[0].Operation)
}

func TestTransaction_Rollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Begin transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// Add an audit entry
	entry := &repository.AuditEntry{
		Timestamp:       time.Now(),
		Operation:       "test_operation",
		Tool:            "test-tool",
		Version:         "1.0.0",
		Status:          "success",
		GpgVerification: "skipped",
	}

	err = tx.AuditRepo().Log(ctx, entry)
	require.NoError(t, err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify data was NOT saved
	repo, _ := sqlite.NewAuditRepository(db.Conn())
	entries, err := repo.Query(ctx, repository.AuditFilter{})
	require.NoError(t, err)
	require.Len(t, entries, 0) // Should be empty because it was rolled back
}

func TestTransaction_MultipleOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Begin transaction
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)

	// 1. Add an audit entry
	auditEntry := &repository.AuditEntry{
		Timestamp:       time.Now(),
		Operation:       "install_test",
		Tool:            "test-tool",
		Version:         "1.0.0",
		Status:          "success",
		GpgVerification: "skipped",
	}
	err = tx.AuditRepo().Log(ctx, auditEntry)
	require.NoError(t, err)

	// 2. Add a cache entry
	err = tx.CacheRepo().Set(ctx, "test-key", []byte("test-data"), time.Hour)
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify all operations succeeded and were saved
	auditRepo, _ := sqlite.NewAuditRepository(db.Conn())
	audits, err := auditRepo.Query(ctx, repository.AuditFilter{})
	require.NoError(t, err)
	require.Len(t, audits, 1)

	cacheRepo, _ := sqlite.NewCacheRepository(db.Conn())
	cacheData, err := cacheRepo.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, []byte("test-data"), cacheData)
}

func TestTransaction_RollbackOnError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// Create wrapper function to simulate error handling with defer
	func() {
		tx, err := tm.Begin(ctx)
		require.NoError(t, err)

		defer func() {
			// Always rollback on error (in this test we just force a rollback)
			_ = tx.Rollback()
		}()

		// Add audit entry
		auditEntry := &repository.AuditEntry{
			Timestamp:       time.Now(),
			Operation:       "install_test",
			Tool:            "test-tool",
			Version:         "1.0.0",
			Status:          "success",
			GpgVerification: "skipped",
		}
		err = tx.AuditRepo().Log(ctx, auditEntry)
		require.NoError(t, err)

		// Simulate an error occurring later in the process
		// The deferred Rollback() will handle cleaning up the audit entry
	}()

	// Verify data was rolled back
	repo, _ := sqlite.NewAuditRepository(db.Conn())
	audits, err := repo.Query(ctx, repository.AuditFilter{})
	require.NoError(t, err)
	require.Len(t, audits, 0)
}

func TestTransaction_Errors(t *testing.T) {
	db, cleanup := setupTestDB(t)

	tm := NewSQLiteTransactionManager(db.Conn())
	ctx := context.Background()

	// 1. Begin success, test Commit and Rollback errors by double calling
	tx, err := tm.Begin(ctx)
	require.NoError(t, err)
	
	err = tx.Commit()
	require.NoError(t, err)
	
	// Double commit should error
	err = tx.Commit()
	require.Error(t, err)

	// Rollback after commit should error
	err = tx.Rollback()
	require.Error(t, err)

	// 2. Test Begin error (db closed)
	cleanup() // This closes the DB

	_, err = tm.Begin(ctx)
	require.Error(t, err)
}
