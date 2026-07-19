# Data Model: Ansible Execution Hook

## Entities

### `installer.PackageInstaller` (Modified)
The package installer struct in `internal/client/installer.go` will be enhanced, but the core struct fields don't need changes. The main logic change is inside its methods.

### `Installer Execution Context` (Concept)
No new persistent database tables are required for this phase. The execution context is passed dynamically as CLI arguments to `ansible-playbook`:
- `app_source_path`: The absolute path where the package was extracted.
- `inventory`: `localhost,`
- `connection`: `local`

## Validations & State Transitions
1. **Pre-condition**: Package downloaded and extracted to `~/.local/share/unistack/<pkg_name>`.
2. **Check**: Does `~/.local/share/unistack/<pkg_name>/app_loader.yml` exist?
   - **If Yes**: State transitions to -> **Ansible Executing**.
     - Runs: `ansible-playbook -i localhost, -c local app_loader.yml -e app_source_path=...`
     - **If Success**: State transitions to -> **Installed**.
     - **If Failure**: State transitions to -> **Failed** (cleanup not strictly required in MVP, but returns non-nil error to CLI).
   - **If No**: State transitions to -> **Installed** (Pure binary/Go logic, no Ansible needed).
