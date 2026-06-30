#!/bin/sh
# =============================================================================
# Base Image Initialization Script (Rocky Variant)
# =============================================================================
# Purpose: Extension point for base image initialization tasks
# Execution: Called by docker-entrypoint.sh during container startup
# Order: Runs after entrypoint.d scripts from the base rocky image
#
# This script is intentionally minimal - it serves as a placeholder for
# future initialization logic specific to the base image layer.
#
# Common use cases:
#   - Additional environment variable setup
#   - Custom directory creation
#   - Service-specific initialization
#   - Health check preparation
#
# Note: The base rocky image already handles:
#   - User creation (PUID/PGID)
#   - Timezone configuration (TZ)
#   - Umask settings (UMASK)
#   - Working directory setup (WORKDIR)
# =============================================================================

# Exit immediately if any command fails
set -e
