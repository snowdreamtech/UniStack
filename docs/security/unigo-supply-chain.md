# unigo Supply Chain Security Analysis

## Overview

This document analyzes the supply chain risks associated with unigo's default registry and provides mitigation strategies.

## Risk Analysis

### 1. Implicit Registry Redirection

**Risk**: unigo's built-in registry can silently redirect tool installations to different backends.

**Example**:

```toml
# You specify:
"github:checkmake/checkmake" = "v0.3.2"

# But unigo's registry maps 'checkmake' to:
aqua:mrtazz/checkmake
```

**Impact**:

- Different maintainer (`mrtazz` vs `checkmake` organization)
- Additional layer (aqua registry) increases attack surface
- Potential for supply chain attacks if registry is comprounigod

### 2. Affected Tools in This Project

Based on unigo registry inspection, the following tools have registry mappings:

```bash
checkmake                     aqua:mrtazz/checkmake
gitleaks                      aqua:gitleaks/gitleaks
hadolint                      aqua:hadolint/hadolint
```

## Mitigation Strategies

### ✅ Already Implemented

1. **Explicit Backend Specification**: All tools in `.unigo.toml` use explicit backends:
   - `github:owner/repo` for GitHub releases
   - `npm:package` for npm packages
   - `pipx:package` for Python packages

2. **Tool Spec Mapping**: In `scripts/lib/lint-wrapper.sh`, we explicitly map tool names to full specs:

   ```bash
   checkmake)
     _UNIRTM_TOOL_SPEC="github:checkmake/checkmake"
     _LINTER_BIN="checkmake"
     ;;
   ```

3. **Version Pinning**: All tools are pinned to specific versions in `.unigo.toml`

### 🔒 Additional Recommendations

#### 1. Disable unigo Registry (Future)

When unigo supports it, consider disabling the default registry:

```toml
[settings]
disable_default_registry = true  # Not yet supported
```

#### 2. Audit Tool Sources

Regularly verify that installed tools match expected sources:

```bash
# Check what unigo actually installed
unigo list

# Verify binary checksums against official releases
unigo exec -- <tool> --version
```

#### 3. Use unigo.lock for Reproducibility

The `unigo.lock` file ensures consistent installations across environments:

```bash
# Verify lock file matches configuration
unigo install --frozen
```

#### 4. Monitor unigo Registry Changes

Watch for changes in unigo's registry that might affect your tools:

```bash
# Check current registry mappings
unigo registry | grep -E "(checkmake|gitleaks|hadolint)"
```

## Verification Steps

### Before Deployment

1. **Verify Tool Sources**:

   ```bash
   unigo list | grep -v "npm:" | grep -v "pipx:"
   ```

2. **Check for Unexpected Backends**:

   ```bash
   unigo list | grep "aqua:"
   ```

   Should only show tools you explicitly configured with aqua backend.

3. **Validate Binary Integrity**:

   ```bash
   # For GitHub releases, verify against official checksums
   unigo where github:checkmake/checkmake
   sha256sum $(unigo where github:checkmake/checkmake)/bin/checkmake
   ```

### During CI/CD

Our CI workflows already implement:

- ✅ Locked unigo versions (`UNIRTM_LOCKED=1`)
- ✅ Explicit tool specs in lint-wrapper.sh
- ✅ Version pinning in .unigo.toml
- ✅ unigo.lock committed to repository

## Related Security Measures

1. **Dependabot**: Monitors unigo tool versions
2. **Trivy**: Scans for vulnerabilities in binaries
3. **SBOM Generation**: Documents all tool dependencies
4. **Signed Commits**: Ensures code integrity

## References

- [unigo Registry Documentation](https://github.com/snowdreamtech/UniGoregistry.html)
- [unigo Security Policy](https://github.com/jdx/unigo/blob/main/SECURITY.md)
- [unigo Paranoid Mode](https://github.com/snowdreamtech/UniGoparanoid)
- [SLSA Framework](https://slsa.dev/)

## Action Items

- [ ] Monitor unigo for registry disable feature
- [ ] Set up automated alerts for registry changes
- [ ] Document tool source verification in CI
- [ ] Consider contributing to unigo for better registry transparency

---

**Last Updated**: 2026-04-16
**Reviewed By**: Security Team
**Next Review**: 2026-07-16
