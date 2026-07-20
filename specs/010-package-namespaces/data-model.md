# Data Model: Package Namespaces

## Package ID Generation

In logical layers and configuration:

- `PackageName`: string (e.g. `snowdreamtech/hello`)

In physical layers (Filesystem, Archive):

- `SafePackageName`: string (e.g. `snowdreamtech_hello`)
- Transformation: `strings.ReplaceAll(PackageName, "/", "_")`
