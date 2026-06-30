#!/bin/sh
# =============================================================================
# Base Image Initialization Script (Debian Variant)
# =============================================================================
# Purpose: Extension point for base image initialization tasks
# Execution: Called by docker-entrypoint.sh during container startup
# Order: Runs after entrypoint.d scripts from the base debian image
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
# Note: The base debian image already handles:
#   - User creation (PUID/PGID)
#   - Timezone configuration (TZ)
#   - Umask settings (UMASK)
#   - Working directory setup (WORKDIR)
#   - DEBIAN_FRONTEND=noninteractive for apt operations
# =============================================================================

# Exit immediately if any command fails
set -e
