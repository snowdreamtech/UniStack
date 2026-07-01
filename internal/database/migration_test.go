package database

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMigrationManager(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_migrations.db")

	ctx := context.Background()
	config := Config{
		Path:    dbPath,
		WALMode: true,
	}

	// Open triggers initialization which runs all migrations
	db, err := Open(ctx, config)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	mgr := NewMigrationManager(db.Conn())

	version, err := mgr.GetCurrentVersion(ctx)
	if err != nil {
		t.Errorf("failed to get current version: %v", err)
	}

	if version != CurrentSchemaVersion {
		t.Errorf("expected version %d, got %d", CurrentSchemaVersion, version)
	}

	// Test rollback
	if err := mgr.Rollback(ctx); err != nil {
		t.Errorf("failed to rollback: %v", err)
	}

	versionAfterRollback, err := mgr.GetCurrentVersion(ctx)
	if err != nil {
		t.Errorf("failed to get version after rollback: %v", err)
	}

	if versionAfterRollback != CurrentSchemaVersion-1 {
		t.Errorf("expected version %d, got %d", CurrentSchemaVersion-1, versionAfterRollback)
	}

	// Apply migrations again
	if err := mgr.ApplyMigrations(ctx); err != nil {
		t.Errorf("failed to re-apply migrations: %v", err)
	}

	finalVersion, err := mgr.GetCurrentVersion(ctx)
	if err != nil {
		t.Errorf("failed to get final version: %v", err)
	}

	if finalVersion != CurrentSchemaVersion {
		t.Errorf("expected final version %d, got %d", CurrentSchemaVersion, finalVersion)
	}
}
