# Quickstart & Validation: refactor-loader

## Prerequisites

- Target container (e.g., Ubuntu 26.04) running in Docker.
- UniStack CLI wrapper available.

## Setup

Start a fresh target container for testing:

```bash
docker run -d --name unistack-target-refactor ubuntu:26.04 sleep infinity
```

## Validation Scenarios

### Scenario 1: Deploy foundation (metapackage)

Run the full foundation scenario and verify it completes without template/file path errors:

```bash
./unistack apply -i "unistack-target-refactor," -c docker ansible/playbooks/scenarios/foundation.yml
```

**Expected Outcome**: The playbook completes successfully. All sub-packages (`mirror`, `repositories`, `sudo`, `user`, `openssh`) are applied.

### Scenario 2: Validate sub-package isolation

Deploy just the `openssh` application to verify it resolves its own defaults and templates:

```bash
./unistack apply -i "unistack-target-refactor," -c docker ansible/playbooks/scenarios/app.yml -e app_name=openssh
```

**Expected Outcome**: The `app` role executes the `openssh` role correctly, templating `/etc/ssh/sshd_config` correctly.

## Cleanup

```bash
docker rm -f unistack-target-refactor
```
