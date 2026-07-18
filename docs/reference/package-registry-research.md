# 🔬 UniStack Package and Registry Design: Deep Dive & Final Solution

> **Note**: This document retains the initial **in-depth research and alternative solutions (Part 1)** for mainstream global package management systems, along with the **finalized implementation plan (Part 2, v3)** decided after multiple rounds of discussion. Both coexist to allow future review of the architectural background and decision logic.

---

# 📖 Part 1: Preliminary Research & Alternative Explorations

> **Goal**: Explore the design space across four dimensions: **Universality, Security, Extensibility, and Speed**.

## 1. Design Anatomy of World-Class Package Managers

We studied 10 of the most successful package management systems to extract their core design wisdom across the four dimensions:

### 📊 Multi-dimensional Comparison Matrix

| System | Package Format | Indexing Strategy | Security Model | Private Registry Support | Speed Rating |
|------|--------|---------|---------|-----------|---------|
| **APT** (Debian) | `.deb` (Binary + Metadata) | Full index `Packages.gz` | GPG Signature + Release file | `sources.list` config | ⭐⭐⭐ |
| **DNF** (RHEL) | `.rpm` (Binary + Metadata) | **SQLite cache** `primary.sqlite` | GPG + repo metadata signature | `.repo` file config | ⭐⭐⭐⭐ |
| **Homebrew** | Ruby Formula (Source description) | **Git repo** + JSON API | Bottle SHA-256 validation | **Tap mechanism** (Any Git repo) | ⭐⭐⭐ |
| **Scoop** (Windows) | **JSON manifest** | Git repo (bucket) | SHA-256 / SHA-512 | Custom bucket (Git repo) | ⭐⭐⭐⭐ |
| **Cargo** (Rust) | `Cargo.toml` (Source description) | **Sparse HTTP Index** (On-demand) | SHA-256 + API token | `[registries]` config | ⭐⭐⭐⭐⭐ |
| **OCI/ORAS** | OCI Manifest (Universal) | Registry API (content-addressable) | **Sigstore/cosign** + SHA-256 | Any OCI-compatible Registry | ⭐⭐⭐⭐ |
| **Nix** | `.nix` derivation (Functional) | Nixpkgs Git monorepo | Content-Addressable (CA) Hash | Flake inputs | ⭐⭐ |
| **mise/aqua** | **YAML manifest** | Git repo + Compile cache | GitHub Release checksum | Custom registry URL | ⭐⭐⭐⭐ |
| **Go modules** | `go.mod` | **GOPROXY protocol** (Simple HTTP) | `go.sum` transparent log | `GOPRIVATE` + `GONOSUMDB` | ⭐⭐⭐⭐⭐ |
| **npm** | `package.json` | CouchDB REST API | `npm audit` + ECDSA signature | `.npmrc` registry URL | ⭐⭐⭐ |

### 🏆 Champions in Each Dimension

#### 1. Universality Champion: OCI (Open Container Initiative)
- ✅ Originally designed for Docker containers, but evolved into a "Universal Distribution Protocol" in v1.1+.
- ✅ Global infrastructure ubiquitous: Docker Hub, GitHub GHCR, AWS ECR, GCP GCR, Harbor.
- ⚠️ Concepts are complex (manifest, layer, config blob), overkill for lightweight "installation maps".

#### 2. Security Champion: TUF (The Update Framework)
- ✅ Designed specifically for software distribution security, protecting even if the Registry is compromised or signing keys are leaked.
- ✅ Defends against rollback attacks, freeze attacks, mix-and-match attacks, etc.
- ⚠️ High implementation complexity, too heavy for smaller projects.

#### 3. Extensibility Champion: Homebrew Tap + aqua Registry
- ✅ Anyone can create their own "Tap" (Git repo) for decentralized distribution.
- ✅ Minimalist format: One YAML/Ruby/JSON file per software.
- ⚠️ Git repo indexing performance degrades significantly with a massive number of packages.

#### 4. Speed Champion: Cargo Sparse Index + Go GOPROXY
- ✅ Avoids downloading full indexes, fetches package info on-demand via lightweight HTTP.
- ✅ Supports cascading proxies and ETag incremental sync.
- ⚠️ Initial full-text search is slower (no complete local index).

---

## 2. Core Design Tensions

1. **Simplicity vs. Expressiveness of Package Format**: JSON (easy to parse) vs. YAML (human-friendly) vs. Ruby/Nix DSL (arbitrary logic). UniStack needs declarative formats; YAML/JSON is sufficient.
2. **Indexing Speed vs. Offline Availability**: Sparse HTTP (fastest online) vs. Git/SQLite full index (offline searchable). UniStack needs a hybrid strategy: online hosting + local SQLite cache.
3. **Security vs. Implementation Cost**: SHA-256 (simplest) vs. GPG vs. Sigstore vs. TUF (most secure). UniStack starts with SHA-256 and reserves architectural space for future upgrades.
4. **Universal Distribution vs. Proprietary Format**: Custom format (total control) vs. OCI Registry/Nix (ecosystem reuse). UniStack leans towards a lightweight, proprietary architecture.

---

## 3. 3 Candidate Solutions from Early Research

### Solution A: Aqua-Inspired — YAML Blueprints + SQLite Local Cache (Score: 17)
- **Concept**: A YAML file describes the installation map, hosted on Git/HTTP, queried locally via SQLite.
- **Pros**: Human-readable, highly extensible, fast parsing, suitable as a carrier for lightweight packages.
- **Cons**: Searching heavily relies on local cache sync.

### Solution B: OCI-Native — Reusing Container Registry Infrastructure (Score: 17)
- **Concept**: Package as an OCI Artifact and push to Docker Hub / GHCR image registries.
- **Pros**: No need to build backend infrastructure; native support for Sigstore signing.
- **Cons**: Slow queries (no native search), relies on complex ORAS client logic.

### Solution C: Nix-Inspired — Content Addressing + Functional Immutability (Score: 13)
- **Concept**: Hash-based immutable storage, perfectly realizing dependency isolation and deduplication.
- **Pros**: Theoretically purest, completely tamper-proof.
- **Cons**: Extremely steep learning curve, infrastructure must be built from scratch, too heavy.

---
---

# 🎯 Part 2: Final Design Solution (v3)

> Based on multiple rounds of decision-making discussions with users, the following pragmatic approach based on **archives + built-in indexing** was finalized.

## 1. Package Format

### 1.1 Standard `.tar.gz` (No Reinventing the Wheel)

```
vim-1.0.0.tar.gz                    # Standard tar.gz, extractable by any tool
│
├── package.yml                      # Package metadata
├── defaults/main.yml                # (Reusing Ansible Role structure)
├── vars/
│   ├── debian.yml                   # System package manager fallback mapping naturally lives here
│   ├── alpine.yml
│   └── ...
├── tasks/main.yml                   # Execution entry
├── templates/
└── files/
```

**Advantage**: Zero learning curve, view contents with `tar xzf`, executed directly by Ansible. Fully reuses the working directory pattern already established in `roles/apps/foundation`.

---

## 2. Package Kinds: `kind: package` vs `kind: meta`

### 2.1 Single Software Package (`kind: package`)

```yaml
# vim-1.0.0/package.yml
apiVersion: "v1"
kind: "package"                     # ⭐ Single software package
name: "vim"
version: "1.0.0"
description: "Vi IMproved text editor"
delivery_mode: "native"
platforms:
  supported: [debian, alpine, redhat, darwin]
```

### 2.2 Meta Package (`kind: meta`)

Similar to `build-essential`, this is a "scenario bundle" comprising dozens of software packages (e.g., foundation).

```yaml
# foundation-1.0.0/package.yml
apiVersion: "v1"
kind: "meta"                        # ⭐ Meta package
name: "foundation"
version: "1.0.0"
description: "System baseline: 70+ essential packages"

# Core: Declarative dependency list
dependencies:
  vim: {}
  curl: {}
  xz: {}
```

**Meta Package Installation Flow**:
1. **Dependency Resolution**: Check the `dependencies` list and find corresponding single software packages in the registry.
2. **Batch Installation**: Parallel download and sequential execution.
3. **Execution Bypass**: The meta package's own tasks are typically empty or bypassed.

---

## 3. Registry Construction: UniStack Native Integration

**The same `unistack` binary acts as both the client and the registry builder.**

### 3.1 Registry Initialization and Management

```bash
# 1. Initialize empty registry
unistack repo init /opt/my-registry

# 2. Add package (Automatically validates format, calculates hash, and places in correct directory)
unistack repo add /opt/my-registry vim-1.0.0.tar.gz

# 3. Build index (Generates SQLite database packages.db)
unistack repo build /opt/my-registry
```

### 3.2 Hosting Methods

**Method A: Static Hosting** (Lowest cost)
Directly host the `/opt/my-registry` directory using Nginx, or `rsync` to S3/GitHub Pages. Because a pre-built SQLite index exists, even a static server allows advanced searching.

**Method B: UniStack Built-in Service**
```bash
unistack repo serve /opt/my-registry --port 8080
```
Starts a built-in lightweight HTTP server, supporting ETag incremental sync and private registry BasicAuth/Bearer authentication.

---

## 4. Registry Directory Structure and URL Rules

```
https://registry.unistack.dev/              # Registry Base URL
│
├── repodata/                                # Index Directory
│   ├── packages.db                          # DNF-style SQLite index
│   ├── packages.db.zst                      # Zstandard compressed version (for client sync)
│   └── repomd.json                          # Registry metadata
│
└── packages/                                # Organized alphabetically
    ├── a/
    │   └── aria2/aria2-1.0.0.tar.gz
    └── f/
        └── foundation/foundation-1.0.0.tar.gz  # Meta package
```

**Download URL** = `Registry Base URL` + `relative_path` (from SQLite)

---

## 5. Security Upgrade Path

The architecture design ensures backward compatibility. Older clients automatically ignore unrecognized new security fields:

1. **v1 (MVP)**: `SHA-256` checksum (Stored in SQLite index + client verification after download).
2. **v2**: `GPG` signed packages and indices.
3. **v3**: `Sigstore cosign` OIDC keyless signing and transparent log auditing.
4. **v4**: `TUF (The Update Framework)` defending against freeze, rollback, and other advanced supply chain attacks.

---

## 6. Embedded Strategy (`go:embed`)

In the Go codebase:
```go
// internal/registry/embedded.go
// Top 50 core foundational packages are compiled directly into the binary, achieving an offline, out-of-the-box, lightning-fast installation experience.
// Only metadata and lightweight configurations are embedded, increasing size by < 500KB.
//
//go:embed builtin/vim builtin/curl builtin/foundation ...
var BuiltinPackages embed.FS
```
The vast long-tail of software packages are retrieved online by syncing the cloud index via `unistack update`.

---

## 7. Architecture Q&A

> **Q1: Since packages are `.tar.gz`, how do I verify their legitimacy?**

**Answer**: We strictly maintain the boundary of the open-source ecosystem:
**Legitimacy verification requires no manipulation inside the file.** Instead, it relies entirely on comparing the `SHA-256` hash provided in the Registry index (`packages.db`). If the hash matches, it proves to be a legitimate, official package.

> **Q2: What format are Ansible Galaxy packages? Do they meet our needs?**

**Answer**: Ansible Galaxy content (Roles and Collections) are essentially standard `.tar.gz` (or just a Git Repo). They **completely fail to meet UniStack's distribution requirements**:
1. **Lack of Native OS Routing**: Galaxy Roles often branch by hardcoding `when: ansible_os_family == 'Debian'`, leading to bloated code. Our `vars/{{ platform }}.yml` fallback mechanism is far more elegant.
2. **Lack of Multiple Delivery Modes**: Galaxy only knows how to run scripts. Our packages, via `delivery_mode` (native/archive/container), allow the Go layer to intelligently choose whether to call the system package manager, download binaries directly, or spin up a container.
3. **Lack of High-Speed Local Indexing**: Galaxy client queries are very slow (heavily reliant on full API network requests). Our DNF mode (SQLite client index) achieves millisecond-level offline and online searches.

---

## 8. Peripheral Ecosystem & Harvester Architecture

To solve the "cold start" dilemma of the initial package management ecosystem, we conceptualized an auxiliary ecosystem project independent of the UniStack core: **`unistack-harvester`**.

### 8.1 "Borrow Rules, Precipitate Data" Strategy
Utilize **Repology (repology.org)**, the industry's most authoritative cross-platform package metadata project, as our initial "Rosetta Stone".
1. **No-Chokehold Design**: Never set the Repology API as a runtime dependency. We periodically download their final Data Dumps or directly parse official source codes (like Debian `Packages`, Alpine `APKINDEX`) to independently generate UniStack's `synonyms.yml` dictionary locally.
2. **Legal and Open-Source Compliance (Clean Room)**: Repology's source rules are under GPL. By "extracting only the computed factual data (facts are not copyrightable)" and open-sourcing our independently written parser scripts under MIT / CC-BY-SA, we perfectly circumvent GPL viral infection.

### 8.2 Ecosystem Evolution Roadmap
* **Phase 1 (Current Best)**: Write minimalist Go scripts, purely parsing external JSON Data Dumps to assemble `package.yml`.
* **Phase 2 (Server-side Isolation)**: If complex Python/Database parsing engines are needed, isolate them entirely within CI/CD (Docker containers), destroying the environment after yielding clean data. Prevent any Python dependencies from creeping into the user-facing Go project.
* **Phase 3 (Native Rule Evolution)**: Once the project scales up, consider developing pure Go-native regex and YAML rule engines (High difficulty, currently lowest priority).
