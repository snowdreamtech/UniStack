# Data Model & Interfaces: refactor-loader

## Ansible Variables (Interfaces)

The sub-packages act as interfaces that consume configuration. Since they are derived from `foundation`, they will inherit or continue to use the variables that were previously scoped to `foundation`.

- `_init_user`: Target username for the `user` sub-package.
- `_init_user_shell`: Target shell for the `user` sub-package.
- `_init_user_password`: Password hash for the user.
- `_init_ssh_port`: Target SSH port for the `openssh` sub-package.
- `_init_permit_root_login`: SSH configuration boolean.
- `_init_password_authentication`: SSH configuration boolean.

## Sub-Package Roles

1. **`mirror`**: Configures package manager mirrors (apt/yum/apk).
2. **`repositories`**: Installs essential core repositories.
3. **`sudo`**: Configures sudoers and privileges.
4. **`user`**: Creates and configures the core deployment user.
5. **`openssh`**: Configures the SSH daemon.

## Metapackage

- **`foundation`**: A wrapper role in `apps/foundation/tasks/main.yml` that invokes:
  1. `apps/mirror`
  2. `apps/repositories`
  3. `apps/sudo`
  4. `apps/user`
  5. `apps/openssh`
