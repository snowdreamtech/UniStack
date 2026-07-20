# Research: refactor-loader

## Loading Contexts in Ansible

**Decision**: Replace `include_tasks` with `include_role` (using `name: "{{ app_name }}"` and optionally `tasks_from: main.yml`).
**Rationale**: `include_tasks` executes tasks within the context of the calling role (`app`). Therefore, any relative paths inside the included tasks (like `src: config.j2` in a `template` task) are resolved against `roles/app/templates/` rather than `roles/apps/openssh/templates/`. Using `include_role` creates a new role execution context, meaning Ansible will correctly look in `roles/apps/openssh/templates/` for files belonging to that package.
**Alternatives considered**: Passing absolute paths to every `template` or `copy` task inside the target package (rejected because it's hacky, verbose, and goes against standard Ansible design).

## Metapackage Implementation

**Decision**: The `foundation` role will become a metapackage. Its `tasks/main.yml` will simply execute a series of `include_role` directives for its dependencies.
**Rationale**: It preserves the simple interface `unistack apply -c docker ansible/playbooks/scenarios/foundation.yml` while utilizing the newly decoupled sub-packages.
**Alternatives considered**: Using `dependencies` in `meta/main.yml` (rejected because it enforces execution before the role's tasks in a rigid way that is sometimes harder to control dynamically via UniStack loaders, and we explicitly control the loading via `app_loader.yml`).

## Splitting Foundation

**Decision**: Extract `openssh`, `sudo`, `repositories`, `mirror`, and `user` into `ansible/roles/apps/`.
**Rationale**: These are distinct configuration domains. For example, some users may want to configure `user` but skip `openssh` if running locally. Modular roles make this possible.
