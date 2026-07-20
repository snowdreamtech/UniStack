# Data Model: Registry Sources

## Source Configuration (JSON)

Stored at `~/.config/unistack/sources.json`.

```json
[
  {
    "name": "default",
    "url": "https://registry.unistack.org"
  },
  {
    "name": "private",
    "url": "http://private.repo"
  }
]
```

- **name**: `string`, unique identifier for the source.
- **url**: `string`, HTTP/HTTPS endpoint pointing to the root of the registry (where `/repodata/repomd.json` can be resolved).

## In-Memory Database View (SQLite)

When reading databases dynamically from multiple sources, the runtime generates a view attached in memory:

```sql
CREATE TEMP VIEW all_packages AS
SELECT *, 'default' as source FROM default_db.packages
UNION ALL
SELECT *, 'private' as source FROM private_db.packages;
```

This dynamic data model means that queries to `all_packages` transparently span all synced registries.
