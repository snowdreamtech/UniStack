# 🇨🇳 **app 角色 (README_zh-CN.md)**

应用 API 抽象层（顶层）

`app` 角色用于定义统一的应用 API，使应用能够在不同容器引擎之间以一致的方式部署。该角色本身不执行任何容器操作，而是负责验证 API、规范化 API，并将其分发给 `container` 角色。

该角色是三层容器平台架构的顶层：

- app（应用 API 抽象层）
    ↓ (通过 app_loader)
- container（应用部署层） / native (软件包/服务层)
- engine_xxx（引擎适配层）
- templates/*_container_loader（底层 loader 层）

## 角色职责

### 1. 验证应用 API

确保必填字段存在，选填字段类型正确。

### 2. 规范化 API 字段

将字段转换为一致的数据结构并应用默认值。

### 3. 分发到应用部署层

将所有规范化后的字段传递给 `container` 角色。

### 4. 审计日志

所有操作通过 `log_loader` 进行结构化审计。

## 变量说明 (Variables)

### 常用变量

| 变量               | 描述                                       | 默认值          |
| ------------------ | ------------------------------------------ | --------------- |
| `app_name`          | 应用名称                                                         | (必填)          |
| `app_delivery_mode` | 部署模式: `container`, `native`, `compose`, `kube`               | (自动探测)      |
| `app_state`         | 期望状态: `present`, `started`, `stopped`, `restarted`, `absent` | `present`       |
| `app_engine`        | 容器引擎: `docker`, `podman`, `containerd`                       | (自动选择)      |
| `app_kube_manifest` | Kubernetes YAML 清单路径                                         | (kube 模式必填) |

## 统一加载器 (Unified Loader: `app_loader.yml`)

在最新的架构中，`app` 角色将 `app_loader.yml` 作为其**唯一的智能引擎**。无论您是正常调用角色还是通过 `tasks_from: app_loader` 调用，它都会执行标准化的 13 步生命周期。

### 1. 高性能探测与短路 (Short-Circuit)

加载器实现了基于**“证据搜集”**的熔断逻辑：

- **探测 (Probing)**：并发检查应用对应的 `vars/` 或 `tasks/` 子目录是否存在。若任一存在则进入 **Module-Mode**；若均不存在则回退至 **Root-Mode**（检查 YAML 文件）。
- **短路 (Short-Circuit)**：如果未发现针对该 `app_name` 的任何自定义定义（无论是目录还是文件），它将跳过后续 8+ 个变量加载和任务解析步骤，直接进入通用分发逻辑。
- **效益**：在保持复杂应用无限灵活性的同时，确保了“零配置”容器部署的最小开销。

### 2. 自动探测逻辑 (Auto-Detection)

根据主机事实自动确定 `app_delivery_mode`（可被显式变量覆盖）：

- `is_k8s_host` -> `kube`。
- `is_docker_host` 或 `is_podman_host` -> `container`。
- 否则 -> `native`。

### 3. 层级变量加载 (Hierarchical Loading)

变量从 `roles/apps/vars/{{ app_name }}/` 中以增量合并方式加载：

1. `default.yml`: 基础元数据。
2. `{{ os_family }}.yml` / `{{ os_distribution }}.yml`: OS 特定覆盖。
3. `{{ app_delivery_mode }}.yml`: 模式特定覆盖（例如：针对物理机 vs 容器使用不同的镜像）。

### 4. 显式变量合并 (Explicit Merging)

支持 defaults 和 overrides 的结构化合并：

- `{{ prefix }}_defaults` + `{{ prefix }}_overrides` => `{{ prefix }}`。
- 默认前缀为 `app_name`。

### 5. 自定义部署任务 (Custom Tasks)

您可以通过提供以下任务文件来完全绕过默认逻辑：

- `tasks/{{ app_name }}/{{ mode }}.yml` 或 `tasks/{{ app_name }}/default.yml`。

如果存在，加载器将执行您的脚本，而不是通用的 `dispatch.yml`。

> [!NOTE]
> `app_engine` 变量（用于容器模式）也支持自动检测，优先使用显式输入，其次降级到探测到的主机运行时。

## 文件结构

tasks/
  main.yml
  validate.yml
  normalize.yml
  dispatch.yml

## 使用示例 (Example Usage)

### 1. 简单场景：零配置容器

无需在任何地方编写 YAML 文件。

```yaml
- include_role:
    name: app
  vars:
    app_name: redis
    app_image: redis:7-alpine
    app_ports: ["6379:6379"]
    app_state: started
```

### 2. 解耦场景：变量驱动

配置存储在 `roles/apps/vars/myapp/default.yml`。

```yaml
- include_role:
    name: app
  vars:
    app_name: myapp
```

### 3. 原生服务：系统软件包

直接部署到宿主机操作系统。

```yaml
- include_role:
    name: app
  vars:
    app_name: nginx
    app_delivery_mode: native
    app_packages: ["nginx-light"]
```

### 4. 自定义场景：特殊逻辑

执行 `roles/apps/tasks/special-app/default.yml` 而不是角色的默认逻辑。

```yaml
- include_role:
    name: app
  vars:
    app_name: special-app
```

### 5. 进阶：多前缀合并

合并多个复杂对象（例如 `mysql` 和 `mysql_metrics`）。

```yaml
- include_role:
    name: app
  vars:
    app_name: mysql
    app_var_prefixes: ["mysql", "mysql_metrics"]
```

## 设计原则

- **可审计 (Auditable)**：每个决策都通过 `APP_LOADER` 和 `APP_DISPATCH` 标签显式记录。
- **可覆盖 (Overridable)**：严格遵守 `defaults` -> `overrides` 模式。
- **高性能 (Performance)**：短路逻辑确保简单应用不会浪费执行周期。
- **引擎无关 (Engine-Agnostic)**：抽象 API 层，与底层 CLI 逻辑解耦。
