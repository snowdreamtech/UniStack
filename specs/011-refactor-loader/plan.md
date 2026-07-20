# Implementation Plan: refactor-loader

**Branch**: `011-refactor-loader` | **Date**: 2026-07-20 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/011-refactor-loader/spec.md`

## Summary

Refactor the `app` loader to use `ansible.builtin.include_role` instead of `ansible.builtin.include_tasks`, solving path resolution issues for templates and files. Concurrently, break the monolithic `foundation` app role into modular sub-packages (`openssh`, `sudo`, `repositories`, `mirror`, `user`) and transform `foundation` into a metapackage.

## Technical Context

**Language/Version**: Ansible 2.14+ (Python 3.9+)

**Primary Dependencies**: `ansible-core`

**Storage**: N/A

**Testing**: Docker-based E2E Playbook tests (ubuntu:26.04, etc.)

**Target Platform**: Linux (Debian/RedHat/Alpine) and macOS

**Project Type**: Infrastructure as Code (Ansible Roles)

**Performance Goals**: N/A (Standard Ansible execution time)

**Constraints**: Must remain idempotent; must cleanly resolve all template/file paths without absolute paths.

**Scale/Scope**: Refactoring the core loader logic and splitting 1 monolithic role into 5 distinct roles + 1 metapackage.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Standards**: Must comply with Ansible best practices (use `include_role` for dynamic role inclusion).
- **Idempotency**: All new modular roles must remain idempotent.
- **Compatibility**: The refactor must not break existing OS support (Linux/macOS).

## Project Structure

### Documentation (this feature)

```text
specs/011-refactor-loader/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (repository root)

```text
ansible/
├── roles/
│   ├── app/
│   │   └── tasks/
│   │       └── app_loader.yml (modified)
│   ├── apps/
│   │   ├── foundation/ (refactored to metapackage)
│   │   ├── mirror/ (new)
│   │   ├── openssh/ (new)
│   │   ├── repositories/ (new)
│   │   ├── sudo/ (new)
│   │   └── user/ (new)
│   └── native/
│       └── tasks/
│           └── native_loader.yml (modified)
```

**Structure Decision**: Splitting `foundation` into flat sub-directories under `apps/` alongside other apps. `foundation` retains its directory but delegates to the new roles.
