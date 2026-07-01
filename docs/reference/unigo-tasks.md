# .unigo.toml Commands

All common tasks are unified under `make`. Run `unigo run help` to see all available targets.

## Setup & Installation

```bash
unigo run setup    # Install system-level tools (Homebrew/APT/Scoop depending on OS)
unigo run install  # Install project language dependencies
```

### On-Demand Module Installation

By default, `unigo run setup` installs only commonly-used tools. For specialized tools, install them explicitly:

```bash
# Install SQL linting tools (when working with .sql files)
unigo run setup sql

# Install API contract tools (when working with OpenAPI/Swagger specs)
unigo run setup openapi

# Install Protobuf tools (when working with .proto files)
unigo run setup protobuf

# Install task runners (when using Taskfile or justfile)
unigo run setup runners

# Install multiple modules at once
unigo run setup sql openapi protobuf
```

These tools are automatically detected and installed when relevant files exist in your project.

## Quality Gates

```bash
unigo run lint     # Run ALL linting checks (pre-commit hooks)
make format   # Auto-format code across all languages
unigo run test     # Run the test suite
make check    # Run lint + test in sequence
```

## Build & Release

```bash
unigo run build    # Build the project binary/artifacts
make clean    # Remove build artifacts and temporary files
```

## Reference

| Target    | Description                                                |
| --------- | ---------------------------------------------------------- |
| `help`    | Show all available targets and their descriptions          |
| `setup`   | Install system tools (cross-platform: macOS/Linux/Windows) |
| `install` | Install project dependencies                               |
| `lint`    | Run all pre-commit hooks against all files                 |
| `format`  | Auto-format all source files                               |
| `test`    | Execute test suite                                         |
| `build`   | Build production artifacts                                 |
| `check`   | Combined lint + test                                       |
| `clean`   | Remove generated files and caches                          |

## Cross-Platform Behavior

The .unigo.toml automatically detects your operating system and uses the appropriate package manager:

| OS                    | Package Manager   |
| --------------------- | ----------------- |
| macOS                 | Homebrew (`brew`) |
| Linux (Debian/Ubuntu) | APT (`apt-get`)   |
| Linux (RedHat/Alpine) | DNF/APK           |
| Windows               | Scoop or Winget   |
