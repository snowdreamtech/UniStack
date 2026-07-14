# 镜像 UID/GID 知识库维护指南

## 📋 目录

1. [知识库概述](#知识库概述)
2. [维护流程](#维护流程)
3. [添加新镜像](#添加新镜像)
4. [验证和测试](#验证和测试)
5. [自动化工具](#自动化工具)
6. [常见问题](#常见问题)

---

## 知识库概述

### 文件位置

```
/roles/container/vars/image_uid_gid_database.yml
```

### 数据结构

知识库包含三个主要部分:

1. **精确匹配** (`image_uid_gid_exact_match`)
   - 用于特定的镜像:标签组合
   - 优先级最高,查询速度最快
   ```yaml
   "mongo:7.0": { uid: "999", gid: "999", source: "..." }
   ```

2. **正则模式匹配** (`image_uid_gid_patterns`)
   - 用于匹配一系列相似镜像
   - 支持版本号通配
   ```yaml
   - regex: "^mongo(db)?:[0-9]+\\.[0-9]+"
     uid: "999"
     gid: "999"
   ```

3. **跳过探测列表** (`image_skip_probe_patterns`)
   - 不应该被探测的镜像(如 distroless)
   - 会直接回退到 root (0:0)

---

## 维护流程

### 🔄 定期维护周期

建议每 **3 个月** 或在以下情况下更新知识库:

- ✅ 新增常用容器应用
- ✅ 主流镜像发布新的主版本
- ✅ 官方镜像 UID/GID 发生变化(极少见)
- ✅ 发现探测失败的镜像

### 📊 维护检查清单

```bash
# 每季度检查清单
□ 检查 Docker Hub 上主流镜像的更新
□ 验证现有映射是否仍然准确
□ 添加团队新引入的镜像
□ 清理已废弃的镜像条目
□ 运行自动化验证脚本
□ 更新 Last Updated 日期
```

---

## 添加新镜像

### 方法 1: 手动查询(推荐)

#### 步骤 1: 验证镜像的 UID/GID

```bash
# 方法 A: 使用 --entrypoint 探测
docker run --rm --entrypoint sh <IMAGE:TAG> -c 'id -u && id -g'

# 示例:
docker run --rm --entrypoint sh postgres:16 -c 'id -u && id -g'
# 输出:
# 999
# 999
```

如果镜像没有 `sh`,尝试:

```bash
# 方法 B: 使用 /bin/sh
docker run --rm --entrypoint /bin/sh <IMAGE:TAG> -c 'id -u && id -g'

# 方法 C: 使用 bash
docker run --rm --entrypoint bash <IMAGE:TAG> -c 'id -u && id -g'
```

#### 步骤 2: 查找官方文档

访问官方仓库确认 UID/GID:

- **Docker Hub**: `https://hub.docker.com/_/<image>`
- **GitHub**: `https://github.com/docker-library/<image>`
- **官方文档**: 查找 Dockerfile 中的 `USER` 指令

#### 步骤 3: 添加到知识库

**精确匹配**(推荐用于常用版本):

```yaml
image_uid_gid_exact_match:
  "postgres:16": { uid: "999", gid: "999", source: "https://github.com/docker-library/postgres" }
  "postgres:15": { uid: "999", gid: "999", source: "https://github.com/docker-library/postgres" }
```

**正则模式**(用于版本系列):

```yaml
image_uid_gid_patterns:
  - regex: "^postgres(ql)?:[0-9]+\\.?[0-9]*"
    uid: "999"
    gid: "999"
    description: "PostgreSQL official images"
```

### 方法 2: 使用自动化脚本

使用自动化工具(见下文)进行批量检测和添加。

---

## 验证和测试

### 手动验证

```bash
# 测试知识库查询是否生效
ansible-playbook -i localhost, test_knowledge_db.yml -e "test_image=mongo:7.0" -v

# 预期输出应包含:
# Method: knowledge_database_exact
# UID: 999
# GID: 999
```

### 批量验证

使用验证脚本检查所有条目:

```bash
cd /Users/snowdream/Workspace/ansible/roles/container/scripts
./validate_image_database.sh
```

### 单元测试

运行完整的测试套件:

```bash
ansible-playbook tests/test_uid_detection.yml --tags knowledge_db -vv
```

---

## 自动化工具

### 🛠️ 工具 1: 镜像 UID/GID 探测器

**位置**: `/roles/container/scripts/detect_image_uid.*`

自动探测指定镜像的 UID/GID:

```bash
# 单个镜像
./detect_image_uid.sh mongo:7.0

# 批量探测
./detect_image_uid.sh mongo:7.0 postgres:16 redis:7-alpine

# 输出格式(YAML)
# "mongo:7.0": { uid: "999", gid: "999", source: "manual_detection" }
```

**可用版本**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

### 🛠️ 工具 2: 知识库验证器

**位置**: `/roles/container/scripts/validate_image_database.*`

验证知识库中所有条目的准确性:

```bash
./validate_image_database.sh

# 输出示例:
# ✓ mongo:7.0 - UID: 999, GID: 999 [PASS]
# ✓ postgres:16 - UID: 999, GID: 999 [PASS]
# ✗ custom:latest - UID mismatch: expected 1000, got 999 [FAIL]
```

**可用版本**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

### 🛠️ 工具 3: 知识库条目生成器

**位置**: `/roles/container/scripts/generate_kb_entry.*`

生成标准格式的知识库条目:

```bash
./generate_kb_entry.sh mongo:7.0 999 999 "https://github.com/docker-library/mongo"

# 输出:
# "mongo:7.0": { uid: "999", gid: "999", source: "https://github.com/docker-library/mongo" }
```

**可用版本**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

---

## 常见问题

### Q1: 如何处理同一镜像的多个版本?

**方案 A**(推荐): 使用正则模式匹配

```yaml
image_uid_gid_patterns:
  - regex: "^mongo:[0-9]+\\.[0-9]+"
    uid: "999"
    gid: "999"
    description: "MongoDB versions 5.0+"
```

**方案 B**: 仅添加常用版本到精确匹配

```yaml
image_uid_gid_exact_match:
  "mongo:latest": { uid: "999", gid: "999", ... }
  "mongo:7.0": { uid: "999", gid: "999", ... }
  "mongo:6.0": { uid: "999", gid: "999", ... }
```

### Q2: 镜像的 UID/GID 在不同版本中可能不同吗?

**答**: 极少见,但确实可能发生:

- **大多数情况**: 官方镜像的 UID/GID 在版本间保持一致
- **例外情况**:
  - Redis Alpine vs Debian 版本(999:1000 vs 999:999)
  - 从 Debian 切换到 Alpine base 的镜像

**处理方法**:

```yaml
# 为不同变种创建单独条目
"redis:7-alpine": { uid: "999", gid: "1000", ... }
"redis:7": { uid: "999", gid: "999", ... }
```

### Q3: 如何处理自定义或私有镜像?

**方案 1**: 添加到知识库(推荐)

```yaml
"mycompany/custom-app:latest": { uid: "1001", gid: "1001", source: "internal" }
```

**方案 2**: 在应用配置中显式指定

```yaml
# roles/apps/vars/myapp/container.yml
myapp_overrides:
  app_volume_owner: "1001"
  app_volume_group: "1001"
```

**方案 3**: 使用模式匹配

```yaml
image_uid_gid_patterns:
  - regex: "^mycompany/.*"
    uid: "1001"
    gid: "1001"
    description: "Company internal images"
```

### Q4: 知识库条目过期怎么办?

**检测过期条目**:

```bash
# 运行验证脚本
./validate_image_database.sh

# 查看失败的条目
# ✗ oldimage:1.0 - Image not found [DEPRECATED]
```

**处理方法**:

1. 删除不再使用的镜像条目
2. 更新 `Last Updated` 日期
3. 添加注释标记废弃版本

```yaml
# 示例: 标记废弃
# DEPRECATED: mongo:3.6 - EOL, use mongo:6.0+
# "mongo:3.6": { uid: "999", gid: "999", ... }
```

### Q5: 正则表达式写错了怎么办?

**测试正则表达式**:

```bash
# 使用 grep 测试
echo "mongo:7.0" | grep -E "^mongo:[0-9]+\.[0-9]+"

# 或使用 Python
python3 -c "import re; print(re.match(r'^mongo:[0-9]+\.[0-9]+', 'mongo:7.0'))"
```

**常见错误**:

```yaml
# ❌ 错误: 未转义点号
regex: "^mongo:[0-9]+.[0-9]+"  # . 匹配任意字符

# ✅ 正确: 转义点号
regex: "^mongo:[0-9]+\\.[0-9]+"  # 只匹配字面上的点
```

### Q6: 如何贡献新条目到知识库?

**内部团队流程**:

1. **创建分支**
   ```bash
   git checkout -b add-kb-entry-kafka
   ```

2. **添加条目**
   ```bash
   # 使用自动化工具
   ./scripts/detect_image_uid.sh confluentinc/cp-kafka:latest

   # 手动编辑知识库
   vim roles/container/vars/image_uid_gid_database.yml
   ```

3. **验证条目**
   ```bash
   ./scripts/validate_image_database.sh
   ```

4. **提交 PR**
   ```bash
   git add roles/container/vars/image_uid_gid_database.yml
   git commit -m "feat: add Kafka image UID/GID mapping"
   git push origin add-kb-entry-kafka
   ```

5. **更新文档**
   - 更新 `Last Updated` 日期
   - 在提交信息中说明来源

---

## 📚 参考资料

### 官方镜像文档

| 镜像 | GitHub | Docker Hub | 默认 UID:GID |
|------|--------|------------|--------------|
| MongoDB | [mongo](https://github.com/docker-library/mongo) | [_/mongo](https://hub.docker.com/_/mongo) | 999:999 |
| PostgreSQL | [postgres](https://github.com/docker-library/postgres) | [_/postgres](https://hub.docker.com/_/postgres) | 999:999 |
| MySQL | [mysql](https://github.com/docker-library/mysql) | [_/mysql](https://hub.docker.com/_/mysql) | 999:999 |
| Redis | [redis](https://github.com/docker-library/redis) | [_/redis](https://hub.docker.com/_/redis) | 999:999 / 999:1000 |
| Nginx | [nginx](https://github.com/nginxinc/docker-nginx) | [_/nginx](https://hub.docker.com/_/nginx) | 101:101 |
| Elasticsearch | [elasticsearch](https://github.com/elastic/elasticsearch) | [elasticsearch](https://hub.docker.com/_/elasticsearch) | 1000:1000 |

### 有用的工具

- **Docker Inspect**: 查看镜像元数据
  ```bash
  docker inspect -f '{{.Config.User}}' IMAGE:TAG
  ```

- **Dive**: 探索镜像层
  ```bash
  dive IMAGE:TAG
  ```

- **Skopeo**: 无需 pull 即可检查镜像
  ```bash
  skopeo inspect docker://IMAGE:TAG
  ```

---

## 📝 更新日志模板

在知识库文件顶部维护更新日志:

```yaml
# =====================================================================
# Last Updated: 2024-01-24
#
# Recent Changes:
#   - 2024-01-24: Initial database creation
#   - 2024-01-24: Added MongoDB Community Server images
#   - 2024-01-24: Added PostgreSQL 14-16 series
# =====================================================================
```

---

## 🔐 安全注意事项

1. **验证来源**: 仅添加官方或受信任的镜像
2. **定期审计**: 每季度检查是否有安全更新
3. **避免敏感信息**: 知识库只包含 UID/GID,不包含密码
4. **版本控制**: 所有更改通过 Git 跟踪

---

## 📞 获取帮助

如有问题,请:

1. 查看此文档的常见问题部分
2. 运行验证脚本诊断问题
3. 检查 Ansible 执行日志(使用 `-vvv`)
4. 联系基础设施团队

---

**维护者**: Infrastructure Team
**最后更新**: 2024-01-24
**版本**: 1.0.0
