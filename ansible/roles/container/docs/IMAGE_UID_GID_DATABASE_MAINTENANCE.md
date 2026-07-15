# Image UID/GID Knowledge Database Maintenance Guide

## 📋 Table of Contents

1. [Knowledge Database Overview](#knowledge-database-overview)
2. [Maintenance Workflow](#maintenance-workflow)
3. [Adding New Images](#adding-new-images)
4. [Validation and Testing](#validation-and-testing)
5. [Automation Tools](#automation-tools)
6. [FAQ](#faq)

---

## Knowledge Database Overview

### File Location

```
/roles/container/vars/image_uid_gid_database.yml
```

### Data Structure

The knowledge database consists of three main sections:

1. **Exact Match** (`image_uid_gid_exact_match`)
   - For specific image:tag combinations
   - Highest priority, fastest lookup

   ```yaml
   "mongo:7.0": { uid: "999", gid: "999", source: "..." }
   ```

2. **Pattern Match** (`image_uid_gid_patterns`)
   - For matching a series of similar images
   - Supports version wildcards

   ```yaml
   - regex: "^mongo(db)?:[0-9]+\\.[0-9]+"
     uid: "999"
     gid: "999"
   ```

3. **Skip Probe List** (`image_skip_probe_patterns`)
   - Images that should not be probed (e.g., distroless)
   - Will fallback to root (0:0)

---

## Maintenance Workflow

### 🔄 Regular Maintenance Cycle

Recommended update frequency: **Every 3 months** or when:

- ✅ Adding new commonly-used container applications
- ✅ Major releases of mainstream images
- ✅ Official images change UID/GID (rare)
- ✅ Probe failures discovered

### 📊 Maintenance Checklist

```bash
# Quarterly checklist
□ Check Docker Hub for mainstream image updates
□ Verify existing mappings are still accurate
□ Add newly introduced team images
□ Clean up deprecated image entries
□ Run automated validation scripts
□ Update Last Updated date
```

---

## Adding New Images

### Method 1: Manual Query (Recommended)

#### Step 1: Verify Image UID/GID

```bash
# Method A: Probe using --entrypoint
docker run --rm --entrypoint sh <IMAGE:TAG> -c 'id -u && id -g'

# Example:
docker run --rm --entrypoint sh postgres:16 -c 'id -u && id -g'
# Output:
# 999
# 999
```

If the image doesn't have `sh`, try:

```bash
# Method B: Use /bin/sh
docker run --rm --entrypoint /bin/sh <IMAGE:TAG> -c 'id -u && id -g'

# Method C: Use bash
docker run --rm --entrypoint bash <IMAGE:TAG> -c 'id -u && id -g'
```

#### Step 2: Find Official Documentation

Visit the official repository to confirm UID/GID:

- **Docker Hub**: `https://hub.docker.com/_/<image>`
- **GitHub**: `https://github.com/docker-library/<image>`
- **Official Docs**: Look for `USER` instruction in Dockerfile

#### Step 3: Add to Knowledge Database

**Exact Match** (recommended for common versions):

```yaml
image_uid_gid_exact_match:
  "postgres:16": { uid: "999", gid: "999", source: "https://github.com/docker-library/postgres" }
  "postgres:15": { uid: "999", gid: "999", source: "https://github.com/docker-library/postgres" }
```

**Pattern Match** (for version series):

```yaml
image_uid_gid_patterns:
  - regex: "^postgres(ql)?:[0-9]+\\.?[0-9]*"
    uid: "999"
    gid: "999"
    description: "PostgreSQL official images"
```

### Method 2: Using Automation Scripts

Use the automation tools (see below) for batch detection and addition.

---

## Validation and Testing

### Manual Validation

```bash
# Test if knowledge database query works
ansible-playbook -i localhost, test_knowledge_db.yml -e "test_image=mongo:7.0" -v

# Expected output should include:
# Method: knowledge_database_exact
# UID: 999
# GID: 999
```

### Batch Validation

Use validation script to check all entries:

```bash
cd /Users/snowdream/Workspace/ansible/roles/container/scripts
./validate_image_database.sh
```

### Unit Tests

Run complete test suite:

```bash
ansible-playbook tests/test_uid_detection.yml --tags knowledge_db -vv
```

---

## Automation Tools

### 🛠️ Tool 1: Image UID/GID Detector

**Location**: `/roles/container/scripts/detect_image_uid.*`

Automatically probe UID/GID for specified images:

```bash
# Single image
./detect_image_uid.sh mongo:7.0

# Batch probe
./detect_image_uid.sh mongo:7.0 postgres:16 redis:7-alpine

# Output format (YAML)
# "mongo:7.0": { uid: "999", gid: "999", source: "manual_detection" }
```

**Available versions**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

### 🛠️ Tool 2: Knowledge Database Validator

**Location**: `/roles/container/scripts/validate_image_database.*`

Validate accuracy of all entries in the knowledge database:

```bash
./validate_image_database.sh

# Example output:
# ✓ mongo:7.0 - UID: 999, GID: 999 [PASS]
# ✓ postgres:16 - UID: 999, GID: 999 [PASS]
# ✗ custom:latest - UID mismatch: expected 1000, got 999 [FAIL]
```

**Available versions**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

### 🛠️ Tool 3: Knowledge Database Entry Generator

**Location**: `/roles/container/scripts/generate_kb_entry.*`

Generate standard-format knowledge database entries:

```bash
./generate_kb_entry.sh mongo:7.0 999 999 "https://github.com/docker-library/mongo"

# Output:
# "mongo:7.0": { uid: "999", gid: "999", source: "https://github.com/docker-library/mongo" }
```

**Available versions**: `.sh` (Shell), `.cmd` (Windows CMD), `.ps1` (PowerShell)

---

## FAQ

### Q1: How to handle multiple versions of the same image?

**Solution A** (Recommended): Use pattern matching

```yaml
image_uid_gid_patterns:
  - regex: "^mongo:[0-9]+\\.[0-9]+"
    uid: "999"
    gid: "999"
    description: "MongoDB versions 5.0+"
```

**Solution B**: Only add common versions to exact match

```yaml
image_uid_gid_exact_match:
  "mongo:latest": { uid: "999", gid: "999", ... }
  "mongo:7.0": { uid: "999", gid: "999", ... }
  "mongo:6.0": { uid: "999", gid: "999", ... }
```

### Q2: Can image UID/GID differ between versions?

**Answer**: Rare, but possible:

- **Most cases**: Official images maintain consistent UID/GID across versions
- **Exceptions**:
  - Redis Alpine vs Debian versions (999:1000 vs 999:999)
  - Images switching from Debian to Alpine base

**Handling**:

```yaml
# Create separate entries for different variants
"redis:7-alpine": { uid: "999", gid: "1000", ... }
"redis:7": { uid: "999", gid: "999", ... }
```

### Q3: How to handle custom or private images?

**Solution 1**: Add to knowledge database (Recommended)

```yaml
"mycompany/custom-app:latest": { uid: "1001", gid: "1001", source: "internal" }
```

**Solution 2**: Explicitly specify in app configuration

```yaml
# roles/apps/vars/myapp/container.yml
myapp_overrides:
  app_volume_owner: "1001"
  app_volume_group: "1001"
```

**Solution 3**: Use pattern matching

```yaml
image_uid_gid_patterns:
  - regex: "^mycompany/.*"
    uid: "1001"
    gid: "1001"
    description: "Company internal images"
```

### Q4: What if knowledge database entries become outdated?

**Detect outdated entries**:

```bash
# Run validation script
./validate_image_database.sh

# View failed entries
# ✗ oldimage:1.0 - Image not found [DEPRECATED]
```

**Handling**:

1. Remove unused image entries
2. Update `Last Updated` date
3. Add comments marking deprecated versions

```yaml
# Example: Mark as deprecated
# DEPRECATED: mongo:3.6 - EOL, use mongo:6.0+
# "mongo:3.6": { uid: "999", gid: "999", ... }
```

### Q5: What if regex pattern is wrong?

**Test regex**:

```bash
# Using grep
echo "mongo:7.0" | grep -E "^mongo:[0-9]+\.[0-9]+"

# Or using Python
python3 -c "import re; print(re.match(r'^mongo:[0-9]+\.[0-9]+', 'mongo:7.0'))"
```

**Common mistakes**:

```yaml
# ❌ Wrong: Unescaped dot
regex: "^mongo:[0-9]+.[0-9]+"  # . matches any character

# ✅ Correct: Escaped dot
regex: "^mongo:[0-9]+\\.[0-9]+"  # Only matches literal dot
```

### Q6: How to contribute new entries?

**Internal team workflow**:

1. **Create branch**

   ```bash
   git checkout -b add-kb-entry-kafka
   ```

2. **Add entry**

   ```bash
   # Use automation tool
   ./scripts/detect_image_uid.sh confluentinc/cp-kafka:latest

   # Manually edit knowledge database
   vim roles/container/vars/image_uid_gid_database.yml
   ```

3. **Validate entry**

   ```bash
   ./scripts/validate_image_database.sh
   ```

4. **Submit PR**

   ```bash
   git add roles/container/vars/image_uid_gid_database.yml
   git commit -m "feat: add Kafka image UID/GID mapping"
   git push origin add-kb-entry-kafka
   ```

5. **Update documentation**
   - Update `Last Updated` date
   - Explain source in commit message

---

## 📚 Reference

### Official Image Documentation

| Image | GitHub | Docker Hub | Default UID:GID |
|-------|--------|------------|-----------------|
| MongoDB | [mongo](https://github.com/docker-library/mongo) | [_/mongo](https://hub.docker.com/_/mongo) | 999:999 |
| PostgreSQL | [postgres](https://github.com/docker-library/postgres) | [_/postgres](https://hub.docker.com/_/postgres) | 999:999 |
| MySQL | [mysql](https://github.com/docker-library/mysql) | [_/mysql](https://hub.docker.com/_/mysql) | 999:999 |
| Redis | [redis](https://github.com/docker-library/redis) | [_/redis](https://hub.docker.com/_/redis) | 999:999 / 999:1000 |
| Nginx | [nginx](https://github.com/nginxinc/docker-nginx) | [_/nginx](https://hub.docker.com/_/nginx) | 101:101 |
| Elasticsearch | [elasticsearch](https://github.com/elastic/elasticsearch) | [elasticsearch](https://hub.docker.com/_/elasticsearch) | 1000:1000 |

### Useful Tools

- **Docker Inspect**: View image metadata

  ```bash
  docker inspect -f '{{.Config.User}}' IMAGE:TAG
  ```

- **Dive**: Explore image layers

  ```bash
  dive IMAGE:TAG
  ```

- **Skopeo**: Inspect images without pulling

  ```bash
  skopeo inspect docker://IMAGE:TAG
  ```

---

## 📝 Changelog Template

Maintain a changelog at the top of the knowledge database file:

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

## 🔐 Security Considerations

1. **Verify sources**: Only add official or trusted images
2. **Regular audits**: Quarterly checks for security updates
3. **Avoid sensitive info**: Knowledge database only contains UID/GID, no passwords
4. **Version control**: All changes tracked via Git

---

## 📞 Getting Help

If you have questions:

1. Check the FAQ section in this document
2. Run validation scripts to diagnose issues
3. Check Ansible execution logs (use `-vvv`)
4. Contact Infrastructure Team

---

**Maintainer**: Infrastructure Team
**Last Updated**: 2024-01-24
**Version**: 1.0.0
