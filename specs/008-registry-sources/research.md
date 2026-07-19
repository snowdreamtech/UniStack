# Phase 0: Research

- Decision: Multi-database approach via SQLite `ATTACH DATABASE`.
- Rationale: Avoids centralizing packages into a single bloated database where removing a source would require deleting hundreds of rows by source name. Instead, dropping a source just means `rm <source_name>.db`.
- Alternatives considered: One massive `packages.db` that tags each package row with the source name. This adds complex migrations and merge logic.

- Decision: Default source behavior.
- Rationale: If `sources.json` does not exist, the `sources` package will mock a default source pointing to `https://registry.unistack.org`. If the user adds their own, the default is preserved unless explicitly removed.
- Alternatives considered: Having no default source. Rejected because it breaks out-of-the-box user experience for zero-config installations.
