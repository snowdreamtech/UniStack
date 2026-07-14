
---

# 🇨🇳 **container/README_zh-CN.md（中文版）**

```markdown

# container 角色

应用部署层（实现层）

`container` 角色负责根据 `app` 角色定义的统一应用 API 来部署应用。
它本身不包含任何容器引擎逻辑，而是将部署请求分发给对应的引擎适配文件：

- engine_docker.yml
- engine_podman.yml
- engine_containerd.yml
- engine_crio.yml
- engine_lxc.yml

该角色是三层容器平台架构的中间层：

app（应用 API 抽象层）
        ↓
container（应用部署层）
        ↓
engine_xxx（引擎适配层）
        ↓
templates/*_container_loader（底层 loader 层）

---

## 角色职责

### 1. 根据应用状态进行路由

根据 `app_state` 决定执行：

- present（创建）
- started（创建并启动）
- restarted（重启）
- stopped（停止）
- absent（删除）

### 2. 透传所有规范化后的 API 字段

来自 `app` 角色的所有字段会被原样传递：

- container_app_name
- container_app_state
- container_app_image
- container_app_env
- container_app_ports
- container_app_volumes
- container_app_networks
- container_app_command
- container_app_args
- container_app_labels
- container_app_restart_policy
- container_app_resources
- container_app_healthcheck

### 3. 分发到对应的引擎适配层

根据变量：

container_engine: docker | podman | containerd | crio | lxc

自动选择对应的 engine_xxx.yml。

### 4. 自动检测容器镜像 UID/GID（仅限 OCI 运行时）

对于 Docker、Podman 和 nerdctl，角色可以自动检测容器镜像内部用户的 UID/GID，
以确保卷的文件所有权正确。

**支持检测的运行时：**

- ✅ Docker（默认）
- ✅ Podman（支持 rootless 模式）
- ✅ nerdctl（containerd CLI）
- ⚠️ LXC/CRI-O（优雅跳过，使用默认所有权）

### 5. 审计日志

所有操作都会通过 `templates` 角色中的 `log_loader` 进行结构化审计。

---

## 文件结构

tasks/
  main.yml
  container_loader.yml
  remove.yml
  engine_docker.yml
  engine_podman.yml
  engine_containerd.yml
  engine_crio.yml
  engine_lxc.yml

---

## 设计原则

- 不写任何引擎逻辑
- 不直接调用任何容器 CLI
- 只做分发，不做实现
- 完全引擎无关
- 完全可审计
- 完全可维护
- 严格职责分离

通过该角色，应用可以在不同容器引擎之间无缝切换，而无需修改应用层 playbook。

---

## 支持的容器运行时

### OCI 兼容运行时（支持 UID/GID 自动检测）

#### Docker（默认）

```yaml
- include_role:
    name: container
    tasks_from: container_loader
  vars:
    container_app_name: nginx
    container_app_image: nginx:alpine
    container_engine: docker  # 可选，这是默认值
```

#### Podman（支持 rootless 模式）

```yaml
- include_role:
    name: container
    tasks_from: container_loader
  vars:
    container_app_name: nginx
    container_app_image: nginx:alpine
    container_engine: podman
```

#### nerdctl（containerd CLI）

```yaml
- include_role:
    name: container
    tasks_from: container_loader
  vars:
    container_app_name: nginx
    container_app_image: nginx:alpine
    container_engine: nerdctl
```

### 其他运行时（不支持 UID/GID 自动检测）

对于 LXC、CRI-O 和其他非 OCI 运行时，UID/GID 检测会被优雅地跳过，
并使用默认的所有权设置。

---

## 使用示例

```yaml
- include_role:
    name: app
  vars:
    app_name: nginx
    app_image: nginx:1.25
    app_ports:
      - "80:80"
    app_state: started
