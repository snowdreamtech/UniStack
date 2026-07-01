// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package database

// Schema contains the SQL schema definitions for all tables
const Schema = `
-- example_items table (placeholder for your generic items)
CREATE TABLE IF NOT EXISTS example_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_example_items_key ON example_items(key);

-- schema_migrations table for tracking migrations (Do not remove)
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    description TEXT NOT NULL
);
`

// CurrentSchemaVersion is the current schema version
const CurrentSchemaVersion = 1

// migrations contains all schema migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Description: "Initial schema",
		Up:          Schema,
		Down:        "DROP TABLE IF EXISTS example_items;",
	},
}
