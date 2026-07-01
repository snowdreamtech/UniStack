package database

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestDatabase_OpenAndInitialize(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	ctx := context.Background()
	config := Config{
		Path:    dbPath,
		WALMode: true,
	}

	db, err := Open(ctx, config)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Errorf("failed to ping db: %v", err)
	}

	version, err := db.GetSchemaVersion(ctx)
	if err != nil {
		t.Errorf("failed to get schema version: %v", err)
	}
	if version != CurrentSchemaVersion {
		t.Errorf("expected version %d, got %d", CurrentSchemaVersion, version)
	}

	// Test basic insert and select
	_, err = db.Conn().ExecContext(ctx, "INSERT INTO example_items (key, value) VALUES (?, ?)", "test_key", "test_value")
	if err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	var val string
	err = db.Conn().QueryRowContext(ctx, "SELECT value FROM example_items WHERE key = ?", "test_key").Scan(&val)
	if err != nil {
		t.Errorf("failed to select: %v", err)
	}
	if val != "test_value" {
		t.Errorf("expected 'test_value', got %q", val)
	}
}

func TestDatabase_Concurrency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_concurrent.db")

	ctx := context.Background()
	config := Config{
		Path:    dbPath,
		WALMode: true,
	}

	db, err := Open(ctx, config)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	var wg sync.WaitGroup
	workers := 50
	operationsPerWorker := 10

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < operationsPerWorker; j++ {
				key := fmt.Sprintf("key_%d_%d", workerID, j)
				value := fmt.Sprintf("val_%d_%d", workerID, j)
				
				// Using immediate transaction logic natively to test locks
				tx, txErr := db.BeginTx(ctx, nil)
				if txErr != nil {
					t.Errorf("worker %d failed to begin tx: %v", workerID, txErr)
					return
				}
				
				_, execErr := tx.ExecContext(ctx, "INSERT INTO example_items (key, value) VALUES (?, ?)", key, value)
				if execErr != nil {
					tx.Rollback()
					t.Errorf("worker %d failed to insert: %v", workerID, execErr)
					return
				}
				
				if commitErr := tx.Commit(); commitErr != nil {
					t.Errorf("worker %d failed to commit: %v", workerID, commitErr)
				}
				
				// Small delay to simulate work
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	var count int
	err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM example_items").Scan(&count)
	if err != nil {
		t.Errorf("failed to count items: %v", err)
	}

	expected := workers * operationsPerWorker
	if count != expected {
		t.Errorf("expected %d items, got %d", expected, count)
	}
}
