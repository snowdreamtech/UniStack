# 📦 UniStack Package Format Specification

This document details the package format specification in the UniStack ecosystem.

---

## 1. Package Format Specification

Packages are primarily distributed to the open-source community, focusing on **high compatibility, high transparency, and zero learning curve**.

### 1.1 Physical Packaging
* **Extension**: **Strictly unified as `.tar.gz`**.
  * *Architectural consideration*: Never use `.zip`. `.tar` perfectly preserves POSIX file permissions natively (like executable `0755`) and Symlinks, which is critical and necessary for infrastructure and code deployment.
* **Compression Algorithm**: Standard gzip.
* **Validity Check and Offline Installation**:
  * **Registry Mode**: When installing via `unistack install <name>`, validity is strictly compared against the `SHA-256` hash provided by the Registry index (`packages.db`).
  * **Local Mode**: To support users installing custom packages via local paths (e.g., `unistack install ./my-tool.tar.gz`), the engine will automatically bypass the Registry check and only print a warning prompt (like `Warning: Installing untrusted local package`), ensuring ultimate openness.

### 1.2 Internal Directory Structure (Ansible Role Native Compatibility)
The uncompressed root directory is the scope of the package, and its internal structure is **strictly compatible with the standard Ansible Role directory tree**. This allows any developer familiar with Ansible to create UniStack packages with zero learning curve.

```text
/ (Archive Root)
├── package.yml          # [Required] UniStack exclusive metadata manifest
├── tasks/
│   └── main.yml         # [Required] Main entry for installation and execution
├── defaults/
│   └── main.yml         # [Optional] Default variables (low priority)
├── vars/
│   ├── main.yml         # [Optional] Internal variables (high priority)
│   ├── debian.yml       # [Optional] OS platform-specific fallback mapping dictionaries
│   └── alpine.yml
├── files/               # [Optional] Static binaries or config files to be distributed
└── templates/           # [Optional] Jinja2 templates (.j2) for dynamic rendering
```

### 1.3 Metadata Specification (`package.yml`)
To ensure robust future extensibility, backward compatibility, and robust handling, `package.yml` discards a flat, simple structure and introduces a strong-typed and layered specification similar to Kubernetes/Helm.

```yaml
# Schema version, reserving compatibility interfaces for potential massive structural refactoring in the future
apiVersion: "v1alpha1"

# Package kind:
#  - package: Standard independent functional package (e.g., vim, nginx)
#  - meta: Meta package, acting as a virtual dependency aggregator (e.g., foundation)
kind: "package" 

# Core metadata
metadata:
  name: "nginx"
  
  # [Recipe Version] (Must be a strict SemVer string):
  # Used only for the UniStack client to determine if the "installation script" itself needs an upgrade (e.g., fixing an Ansible bug).
  # Note: Must use a string (e.g., "1.0.1") to prevent YAML float parsing traps.
  # [Tech Selection]: The Go backend strictly uses `github.com/Masterminds/semver/v3` for version parsing,
  # perfectly accommodating versions with or without the 'v' prefix.
  version: "1.0.1"
  
  # [App Version / Real Software Version]:
  # To completely eliminate user confusion, the real software version is separated out here.
  # Given the severe fragmentation of built-in versions across Linux distros (e.g., Debian has 1.18, Alpine has 1.24),
  # this field can not only be a single string but natively supports arrays or SemVer range syntax!
  # [Tech Selection]: Thanks to the powerful Constraint engine in `Masterminds/semver/v3`,
  # the Go client can natively parse and process complex dependency range constraints.
  appVersion: 
    - "1.18.x"
    - "1.24.x"
    # Or write as a single string range: ">= 1.18.0, < 1.25.0"
    
  description: "High performance web server"
  authors: ["SnowdreamTech <snowdreamtech@qq.com>"]
  homepage: "https://nginx.org"
  license: "MIT"
  tags: ["web", "infrastructure"]

# Normalized environment support declaration (Precise error prevention mechanism)
compatibility:
  # Uses an array of objects to precisely express N×M support relationships.
  # Rule: 'os' and 'arch' must strictly be [lowercase]!
  - os: "debian"
    # [Optional] Constrain the minimum major version of the underlying OS (e.g., if a package is available in Debian 13 but not 12)
    min: "13"
    # If 'archs' is omitted, it means all architectures under this OS are supported
    archs: ["amd64", "arm64"]
  - os: "alpine"
    # Assuming only amd64 compiled packages exist for Alpine
    archs: ["amd64"]
  - os: "redhat"
    min: "9"

# Dependency declaration (Dictionary structure, retaining the ability to expand version constraints in the future)
# !!! CRITICALLY IMPORTANT: All dependencies declared here refer to package names within the UniStack ecosystem, NOT underlying Linux packages!
# For example, if 'foundation' is depended upon, the engine will first download foundation's .tar.gz from the UniStack Registry and execute its Ansible tasks.
dependencies:
  # Current version only writes the package name, future expansions can seamlessly support `foundation: { version: ">= 1.0.0" }`
  foundation: {}
  curl: {}
```

### 1.4 Meta Package Specification (Meta Package)
A Meta package (`kind: meta`) is a special specification in the ecosystem used for "one-click installation bundles" (e.g., installing `foundation` automatically installs vim, git, curl, etc.).

**Special Constraints for Meta Packages:**
1. **Execution Bypass**: When UniStack identifies a `kind: meta`, it **will NOT execute** the `tasks/main.yml` in that package. Its existence is purely to declare dependencies.
2. **Dependency Unrolling**: The engine pushes the meta package's `dependencies` field into the installation queue and uses Topology Sort to resolve and install all sub-packages first.
3. **Minimalist Directory**: A meta package's `.tar.gz` usually contains only a barebones `package.yml`, without massive Ansible Task and Files directories, making it extremely lightweight.

```yaml
# foundation/package.yml (Meta package example)
apiVersion: "v1alpha1"
kind: "meta"

# Meta packages share the exact same metadata structure as standard packages! (Even if left blank, Schema constraints still apply)
metadata:
  name: "foundation"
  version: "1.0.0"
  description: "The core foundation toolchain for all Linux nodes"
  authors: ["SnowdreamTech <snowdreamtech@qq.com>"]
  homepage: "https://github.com/snowdreamtech/unistack"
  license: "MIT"

# Dependencies here all point to software packages within the UniStack ecosystem
dependencies:
  vim: {}
  curl: {}
  git: {}
  jq: {}
```
