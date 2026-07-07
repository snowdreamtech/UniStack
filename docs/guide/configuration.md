# Configuration

## Project Hydration

After cloning from the template, run the hydration script to replace all placeholder values:

```bash
bash scripts/init-project.sh
```

The script walks you through:

| Placeholder     | Example          |
| --------------- | ---------------- |
| `template`      | `my-awesome-app` |
| `snowdreamtech` | `myorg`          |
| `snowdream`     | `johndoe`        |

It replaces these strings throughout all files and optionally re-initializes the Git repository.

## Environment Variables

The project uses `.env` files for local configuration (excluded from Git via `.gitignore`):

```bash
cp .env.example .env  # if .env.example exists
```

## Pre-commit Configuration

Pre-commit hooks are defined in `.pre-commit-config.yaml`. To disable a hook temporarily:

```yaml
# Comment out the hook
# - id: hadolint
```

To skip hooks for a single commit (emergency use only):

```bash
git commit --no-verify -m "chore: emergency fix"
```

## UniStack Tool Manager Configuration

The project uses [unistack](https://github.com/snowdreamtech/UniStack) for managing development tools. Configuration is in `.unistack.toml`.

### Important Security Requirements

**🛡️ Aqua Registry is DISABLED** for supply chain security:

```toml
[settings]
# Aqua Registry repackages binaries and loses attestations
aqua.baked_registry = false
aqua.github_attestations = false
aqua.slsa = false
aqua.cosign = false
aqua.minisign = false
```

**✅ ASDF compatibility is ENABLED** for file format support:

```toml
[settings]
# Allows reading .tool-versions files (format only, not related to Aqua)
asdf_compat = true
```

For detailed unistack configuration guidelines, see:

- [UniStack Configuration Best Practices](../reference/unistack-configuration.md)
- [UniStack Attestation Error Troubleshooting](../troubleshooting/unistack-attestation-error.md)
