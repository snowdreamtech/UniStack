# Quickstart: Registry Sources Validation

This guide explains how to validate the client-side sources feature from end-to-end.

## Prerequisites

1. Have `unistack` compiled locally.
2. Have the CLI initialized (run `unistack source list` to confirm it prints the default source).

## Validation Scenarios

### 1. Add and List Sources

```bash
# Verify initial state
go run main.go source list
# Output should show 'default' -> 'https://registry.unistack.org'

# Add a custom source
go run main.go source add custom http://localhost:8080

# List again to confirm addition
go run main.go source list
```

### 2. Update a Source

```bash
# Update the URL of the 'custom' source
go run main.go source update custom http://custom-registry.com

# Verify the URL changed
go run main.go source list
```

### 3. Sync Multiple Sources

```bash
# Run the registry update command
go run main.go update

# Verify that both default.db and custom.db are created in the cache
ls -l ~/.cache/unistack/registry/
```

### 4. Remove a Source

```bash
# Remove the custom source
go run main.go source remove custom

# Verify it was removed from configuration
go run main.go source list

# Verify the local DB cache was deleted
ls -l ~/.cache/unistack/registry/custom.db # Should fail
```
