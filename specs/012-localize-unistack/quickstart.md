# Quickstart: Localize UniGo to UniStack

**Phase 1 Output**

## Validation Scenarios

1. Verify package names in PyPI metadata:

   ```bash
   cat pypi/pyproject.toml | grep name
   # Expected: name = "snowdreamtech-unistack"
   ```

2. Validate build scripts use `unistack`:

   ```bash
   grep unigo pypi/scripts/build.sh
   # Expected: No output
   ```

3. Validate workflow action:

   ```bash
   cat .github/workflows/goreleaser.yml | grep proxy.golang.org
   # Expected: https://proxy.golang.org/github.com/snowdreamtech/unistack/...
   ```
