# Quickstart Validation: Package Namespaces

1. **Modify a Package**: Set `name: snowdreamtech/hello` in `package.yml`.
2. **Build Registry**: Run `./unistack registry build local_repo local_repo`.
   - **Expectation**: Tarball is created at `local_repo/packages/s/snowdreamtech_hello-1.0.0.tar.gz`.
3. **Install Package**: Run `./unistack install snowdreamtech/hello`.
   - **Expectation**: Extracts to `~/.local/share/unistack/packages/snowdreamtech_hello-1.0.0` and completes successfully.
