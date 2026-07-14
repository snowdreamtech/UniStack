# app Role

Application API Layer (Abstract Layer)

The `app` role defines a unified application API for deploying containerized applications across multiple container engines. It does not perform any container operations itself. Instead, it validates and normalizes the API fields, then dispatches to the `container` role for actual deployment.

This role forms the top layer of the three‑layer container platform:

app (Application API Layer)
        ↓ (via app_loader)
container (Application Deployment Layer) / native (Package/Service Layer)
        ↓
engine_xxx (Engine Adapter Layer)
        ↓
templates/*_container_loader (Low‑level Loader Layer)

## Responsibilities

### 1. Validate application API

Ensures required fields exist and optional fields follow type rules.

### 2. Normalize API fields

Converts fields into consistent structures and applies defaults.

### 3. Dispatch to container deployment layer

Forwards all normalized fields to the `container` role.

### 4. Provide audit logging

All operations are logged using the shared `log_loader`.

## Variables

### Shared Variables

| Variable            | Description                                | Default             |
| ------------------- | ------------------------------------------ | ------------------- |
| `app_name`          | Name of the application                    | (required)          |
| `app_delivery_mode` | `container`, `native`, `compose`, or `kube` | (auto-detected)     |
| `app_state`         | `present`, `started`, `stopped`, `restarted`, `absent` | `present` |
| `app_engine`        | Container engine (`docker`, `podman`, `containerd`) | (auto-selected) |
| `app_kube_manifest` | Path to Kubernetes YAML manifest           | (required for kube) |

## The Unified Loader (`app_loader.yml`)

As of the latest architecture, the `app` role uses `app_loader.yml` as its **exclusive intelligence engine**. Whether you call the role normally or via `tasks_from: app_loader`, it executes a standardized 13-step lifecycle.

### 1. High-Performance Discovery

The loader implements an **"Evidence-based Short-Circuit"** logic:

- **Probing**: It concurrently checks for the existence of application subdirectories (`vars/myapp/` or `tasks/myapp/`). If either exists, it enters **Module-Mode**; otherwise, it falls back to **Root-Mode** (probing for YAML files).
- **Short-Circuit**: If no custom definitions (folders or files) are found for the `app_name`, it skips the subsequent 8+ variable loading and task resolution steps, proceeding directly to the generic dispatch logic.
- **Benefit**: Minimum overhead for "Zero-Config" container deployments while maintaining infinite flexibility for complex apps.

### 2. Auto-Detection Logic

It automatically determines `app_delivery_mode` based on host facts (can be overridden):

- `is_k8s_host` -> `kube`.
- `is_docker_host` or `is_podman_host` -> `container`.
- Otherwise -> `native`.

### 3. Hierarchical Variable Loading (Additive)

Metadata is loaded from `roles/apps/vars/{{ app_name }}/` with the following priority (Additive):

1. `default.yml`: Base metadata.
2. `{{ os_family }}.yml` / `{{ os_distribution }}.yml`: OS-specific overrides.
3. `{{ app_delivery_mode }}.yml`: Mode-specific overrides (e.g., specific images for `native` vs `container`).

### 4. Explicit Variable Merging

Supports structured merging of defaults and overrides:

- `{{ prefix }}_defaults` + `{{ prefix }}_overrides` => `{{ prefix }}`.
- Default prefix is `app_name`. Override via `app_var_prefix` or `app_var_prefixes`.

### 5. Custom Deployment Tasks

You can completely bypass the default role logic by providing a task file at:

- `tasks/{{ app_name }}/{{ mode }}.yml` or `tasks/{{ app_name }}/default.yml`.

If found, the loader executes your script instead of the generic `dispatch.yml`.

> [!NOTE]
> The `app_engine` variable (used in container modes) also supports auto-detection, preferring explicit user input then falling back to the detected host runtime.

## File Structure

tasks/
  main.yml
  validate.yml
  normalize.yml
  dispatch.yml

## Example Usage

### 1. Simple Case: Zero-Config Container

No YAML files needed anywhere else.

```yaml
- include_role:
    name: app
  vars:
    app_name: redis
    app_image: redis:7-alpine
    app_ports: ["6379:6379"]
    app_state: started
```

### 2. Decoupled Case: Variable-Driven

Configuration stored in `roles/apps/vars/myapp/default.yml`.

```yaml
- include_role:
    name: app
  vars:
    app_name: myapp
```

### 3. Native Service: OS Package

Deploy direct to the host OS.

```yaml
- include_role:
    name: app
  vars:
    app_name: nginx
    app_delivery_mode: native
    app_packages: ["nginx-light"]
```

### 4. Custom Case: Specialized Logic

Execute `roles/apps/tasks/special-app/default.yml` instead of the role's default logic.

```yaml
- include_role:
    name: app
  vars:
    app_name: special-app
```

### 5. Advanced: Multi-Prefix Merging

Merge multiple complex objects (e.g. `mysql` and `mysql_metrics`).

```yaml
- include_role:
    name: app
  vars:
    app_name: mysql
    app_var_prefixes: ["mysql", "mysql_metrics"]
```

## Design Principles

- **Auditable**: Every decision is explicitly logged with `APP_LOADER` and `APP_DISPATCH` labels.
- **Overridable**: Strict adherence to the `defaults` -> `overrides` pattern.
- **Performance**: Short-circuit logic ensures no wasted cycles for simple apps.
- **Engine-Agnostic**: Abstract API layer decoupled from Low-level CLI logic.
