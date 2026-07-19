# Quickstart: Local Package Installation

## Prerequisites

- A dummy `.tar.gz` package containing an executable file.
- The `unistack` binary compiled.

## Testing Setup

Create a dummy package:
```bash
mkdir -p dummy_pkg/bin
echo '#!/bin/bash' > dummy_pkg/bin/hello
echo 'echo "Hello from unistack package!"' >> dummy_pkg/bin/hello
chmod +x dummy_pkg/bin/hello
tar -czvf hello-1.0.0.tar.gz -C dummy_pkg .
rm -rf dummy_pkg
```

## Scenario 1: Local File Installation

1. Run the install command against the local file:
   ```bash
   go run main.go install ./hello-1.0.0.tar.gz
   ```
2. Verify the package is extracted:
   ```bash
   ls ~/.local/share/unistack/packages/hello-1.0.0/bin/hello
   ```
3. Verify the symlink exists:
   ```bash
   ls -l ~/.local/bin/hello
   ```
4. Run the installed package:
   ```bash
   ~/.local/bin/hello
   # Expected Output: Hello from unistack package!
   ```

## Scenario 2: Remote Installation (Integration)

Assuming the package is registered in `packages.db` and available on the local test server:

1. Run the remote install command:
   ```bash
   go run main.go install hello
   ```
2. Verify it downloads, extracts, and links successfully as in Scenario 1.
