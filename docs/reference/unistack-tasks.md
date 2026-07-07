# .unistack.toml Commands

All common tasks are unified under `make`. Run `unistack run help` to see all available targets.

## Setup & Installation

```bash
unistack run setup    # Install system-level tools (Homebrew/APT/Scoop depending on OS)
unistack run install  # Install project language dependencies
```

### On-Demand Module Installation

By default, `unistack run setup` installs only commonly-used tools. For specialized tools, install them explicitly:

```bash
# Install SQL linting tools (when working with .sql files)
unistack run setup sql

# Install API contract tools (when working with OpenAPI/Swagger specs)
unistack run setup openapi

# Install Protobuf tools (when working with .proto files)
unistack run setup protobuf

# Install task runners (when using Taskfile or justfile)
unistack run setup runners

# Install multiple modules at once
unistack run setup sql openapi protobuf
```

These tools are automatically detected and installed when relevant files exist in your project.

## Quality Gates

```bash
unistack run lint     # Run ALL linting checks (pre-commit hooks)
make format   # Auto-format code across all languages
unistack run test     # Run the test suite
make check    # Run lint + test in sequence
```

## Build & Release

```bash
unistack run build    # Build the project binary/artifacts
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

The .unistack.toml automatically detects your operating system and uses the appropriate package manager:

| OS                    | Package Manager   |
| --------------------- | ----------------- |
| macOS                 | Homebrew (`brew`) |
| Linux (Debian/Ubuntu) | APT (`apt-get`)   |
| Linux (RedHat/Alpine) | DNF/APK           |
| Windows               | Scoop or Winget   |
