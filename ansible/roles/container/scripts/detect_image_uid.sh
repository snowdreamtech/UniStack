#!/bin/sh
# =====================================================================
# detect_image_uid.sh
#
# Purpose:
#   Automatically detect UID/GID for container images.
#   Outputs results in YAML format for easy addition to knowledge database.
#
# Compatibility:
#   - POSIX sh compliant (no bash-specific features)
#   - macOS, Linux, Windows (Git Bash/WSL/MSYS2)
#
# Usage:
#   ./detect_image_uid.sh <image:tag> [image:tag ...]
#
# Examples:
#   ./detect_image_uid.sh mongo:7.0
#   ./detect_image_uid.sh mongo:7.0 postgres:16 redis:7-alpine
# =====================================================================

# Ensure POSIX behavior
set -eu

# =====================================================================
# Cross-platform compatibility helpers
# =====================================================================

# Detect OS
detect_os() {
  case "$(uname -s)" in
  Darwin*) echo "macos" ;;
  Linux*) echo "linux" ;;
  CYGWIN* | MINGW* | MSYS*) echo "windows" ;;
  *) echo "unknown" ;;
  esac
}

OS_TYPE="$(detect_os)"

# Check if command exists (POSIX-compliant)
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Check if running in color-capable terminal
supports_color() {
  if [ -t 1 ]; then
    # Check if TERM is set and not "dumb"
    if [ -n "${TERM:-}" ] && [ "$TERM" != "dumb" ]; then
      return 0
    fi
  fi
  return 1
}

# Color codes (only if terminal supports it)
if supports_color; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  NC='\033[0m'
else
  RED=''
  GREEN=''
  YELLOW=''
  BLUE=''
  NC=''
fi

# =====================================================================
# Output functions
# =====================================================================

print_info() {
  printf "${BLUE}[INFO]${NC} %s\n" "$*"
}

print_success() {
  printf "${GREEN}[OK]${NC} %s\n" "$*"
}

print_warning() {
  printf "${YELLOW}[WARN]${NC} %s\n" "$*"
}

print_error() {
  printf "${RED}[ERROR]${NC} %s\n" "$*" >&2
}

# =====================================================================
# Docker compatibility check
# =====================================================================

check_docker() {
  if ! command_exists docker; then
    print_error "Docker is not installed or not in PATH"
    print_info "Please install Docker: https://docs.docker.com/get-docker/"
    return 1
  fi

  # Test if Docker daemon is running
  if ! docker info >/dev/null 2>&1; then
    print_error "Docker daemon is not running"
    print_info "Please start Docker Desktop or Docker service"
    return 1
  fi

  return 0
}

# =====================================================================
# String manipulation (POSIX-compliant)
# =====================================================================

# Extract first line from output
get_first_line() {
  printf "%s" "$1" | sed -n '1p' | tr -d '\r\n'
}

# Extract second line from output
get_second_line() {
  printf "%s" "$1" | sed -n '2p' | tr -d '\r\n'
}

# Check if string is numeric
is_numeric() {
  case "$1" in
  '' | *[!0-9]*) return 1 ;;
  *) return 0 ;;
  esac
}

# Extract image name without tag
get_image_name() {
  printf "%s" "$1" | sed 's/:.*//'
}

# Check if image has registry prefix
has_registry() {
  case "$1" in
  */*) return 0 ;;
  *) return 1 ;;
  esac
}

# =====================================================================
# Docker operations
# =====================================================================

# Pull image if not present
pull_image_if_needed() {
  image="$1"

  if docker image inspect "$image" >/dev/null 2>&1; then
    return 0
  fi

  print_info "Pulling image: $image"
  if docker pull "$image" >/dev/null 2>&1; then
    return 0
  else
    print_error "Failed to pull image: $image"
    return 1
  fi
}

# Try to detect UID/GID using docker run
try_docker_run() {
  image="$1"
  entrypoint="$2"

  result=$(docker run --rm --entrypoint "$entrypoint" "$image" -c 'id -u && id -g' 2>/dev/null || true)

  if [ -n "$result" ]; then
    uid=$(get_first_line "$result")
    gid=$(get_second_line "$result")

    if is_numeric "$uid" && is_numeric "$gid"; then
      printf "%s:%s" "$uid" "$gid"
      return 0
    fi
  fi

  return 1
}

# Try to detect UID/GID using docker inspect
try_docker_inspect() {
  image="$1"

  user_info=$(docker inspect -f '{{.Config.User}}' "$image" 2>/dev/null || true)

  if [ -z "$user_info" ]; then
    return 1
  fi

  # Parse UID:GID or UID format
  case "$user_info" in
  *:*)
    uid=$(printf "%s" "$user_info" | cut -d: -f1)
    gid=$(printf "%s" "$user_info" | cut -d: -f2)
    ;;
  *)
    uid="$user_info"
    gid="$user_info"
    ;;
  esac

  if is_numeric "$uid" && is_numeric "$gid"; then
    printf "%s:%s" "$uid" "$gid"
    return 0
  fi

  return 1
}

# =====================================================================
# Source URL detection
# =====================================================================

generate_source_url() {
  image="$1"

  if has_registry "$image"; then
    # Image with registry/namespace
    image_without_tag=$(get_image_name "$image")
    printf "https://hub.docker.com/r/%s" "$image_without_tag"
  else
    # Official Docker library image
    image_name=$(get_image_name "$image")
    printf "https://github.com/docker-library/%s" "$image_name"
  fi
}

# =====================================================================
# Main detection logic
# =====================================================================

detect_image_uid_gid() {
  image="$1"
  uid=""
  gid=""
  method=""

  print_info "Detecting UID/GID for: $image"

  # Pull image if needed
  if ! pull_image_if_needed "$image"; then
    return 1
  fi

  # Method 1: Try --entrypoint sh
  if [ -z "$uid" ]; then
    if result=$(try_docker_run "$image" "sh"); then
      uid=$(printf "%s" "$result" | cut -d: -f1)
      gid=$(printf "%s" "$result" | cut -d: -f2)
      method="--entrypoint sh"
      print_success "Detected via --entrypoint sh: UID=$uid, GID=$gid"
    fi
  fi

  # Method 2: Try --entrypoint /bin/sh
  if [ -z "$uid" ]; then
    if result=$(try_docker_run "$image" "/bin/sh"); then
      uid=$(printf "%s" "$result" | cut -d: -f1)
      gid=$(printf "%s" "$result" | cut -d: -f2)
      method="--entrypoint /bin/sh"
      print_warning "Detected via --entrypoint /bin/sh: UID=$uid, GID=$gid"
    fi
  fi

  # Method 3: Try --entrypoint bash
  if [ -z "$uid" ]; then
    if result=$(try_docker_run "$image" "bash"); then
      uid=$(printf "%s" "$result" | cut -d: -f1)
      gid=$(printf "%s" "$result" | cut -d: -f2)
      method="--entrypoint bash"
      print_warning "Detected via --entrypoint bash: UID=$uid, GID=$gid"
    fi
  fi

  # Method 4: Try docker inspect
  if [ -z "$uid" ]; then
    if result=$(try_docker_inspect "$image"); then
      uid=$(printf "%s" "$result" | cut -d: -f1)
      gid=$(printf "%s" "$result" | cut -d: -f2)
      method="docker inspect"
      print_warning "Detected via docker inspect: UID=$uid, GID=$gid"
    fi
  fi

  # Validate results
  if [ -z "$uid" ] || [ -z "$gid" ]; then
    print_error "Failed to detect UID/GID for: $image"
    return 1
  fi

  # Generate source URL
  source_url=$(generate_source_url "$image")

  # Output YAML format
  printf "\n"
  printf "  \"%s\": { uid: \"%s\", gid: \"%s\", source: \"%s\" }\n" "$image" "$uid" "$gid" "$source_url"

  print_success "Generated YAML entry for: $image (method: $method)"
  return 0
}

# =====================================================================
# Date formatting (cross-platform)
# =====================================================================

get_current_date_utc() {
  # Try GNU date first (Linux)
  if date -u +"%Y-%m-%d %H:%M:%S UTC" 2>/dev/null; then
    return 0
  fi

  # Try BSD date (macOS)
  if date -u +"%Y-%m-%d %H:%M:%S UTC" 2>/dev/null; then
    return 0
  fi

  # Fallback
  printf "%s\n" "$(date)"
}

# =====================================================================
# Usage information
# =====================================================================

show_usage() {
  cat <<'EOF'
Usage: detect_image_uid.sh <image:tag> [image:tag ...]

Detects UID/GID for container images and outputs YAML format.

Arguments:
  image:tag     One or more container images with tags

Examples:
  detect_image_uid.sh mongo:7.0
  detect_image_uid.sh mongo:7.0 postgres:16 redis:7-alpine

Output:
  YAML-formatted entries ready to copy to:
  roles/container/vars/image_uid_gid_database.yml

Requirements:
  - Docker must be installed and running
  - Network access to pull images (if not cached)

Compatibility:
  - macOS, Linux, Windows (Git Bash/WSL/MSYS2)
  - POSIX sh compliant
EOF
}

# =====================================================================
# Main execution
# =====================================================================

main() {
  # Check arguments
  if [ $# -eq 0 ]; then
    show_usage
    exit 1
  fi

  # Check Docker availability
  if ! check_docker; then
    exit 1
  fi

  # Print header
  printf "# ==============================================\n"
  printf "# Image UID/GID Detection Results\n"
  printf "# Generated: %s\n" "$(get_current_date_utc)"
  printf "# Platform: %s\n" "$OS_TYPE"
  printf "# ==============================================\n"
  printf "\n"
  printf "# Add these entries to image_uid_gid_exact_match:\n"

  # Process each image
  success_count=0
  fail_count=0

  for image in "$@"; do
    if detect_image_uid_gid "$image"; then
      success_count=$((success_count + 1))
    else
      fail_count=$((fail_count + 1))
    fi
    printf "\n"
  done

  # Print summary
  printf "\n"
  printf "# ==============================================\n"
  printf "# Summary: %d succeeded, %d failed\n" "$success_count" "$fail_count"
  printf "# ==============================================\n"

  if [ "$fail_count" -eq 0 ]; then
    exit 0
  else
    exit 1
  fi
}

# Run main function
main "$@"
