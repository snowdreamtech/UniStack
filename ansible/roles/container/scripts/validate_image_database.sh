#!/bin/sh
# =====================================================================
# validate_image_database.sh
#
# Purpose:
#   Validate all entries in the image UID/GID knowledge database.
#   Checks if the stored UID/GID matches actual container values.
#
# Compatibility:
#   - POSIX sh compliant (no bash-specific features)
#   - macOS, Linux, Windows (Git Bash/WSL/MSYS2)
#
# Usage:
#   ./validate_image_database.sh
# =====================================================================

set -eu

# =====================================================================
# Cross-platform compatibility helpers
# =====================================================================

detect_os() {
  case "$(uname -s)" in
  Darwin*) echo "macos" ;;
  Linux*) echo "linux" ;;
  CYGWIN* | MINGW* | MSYS*) echo "windows" ;;
  *) echo "unknown" ;;
  esac
}

OS_TYPE="$(detect_os)"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

supports_color() {
  if [ -t 1 ]; then
    if [ -n "${TERM:-}" ] && [ "$TERM" != "dumb" ]; then
      return 0
    fi
  fi
  return 1
}

if supports_color; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  NC='\033[0m'
else
  RED=''
  GREEN=''
  YELLOW=''
  NC=''
fi

# =====================================================================
# Output functions
# =====================================================================

print_success() {
  printf "${GREEN}[PASS]${NC} %s\n" "$*"
}

print_warning() {
  printf "${YELLOW}[SKIP]${NC} %s\n" "$*"
}

print_error() {
  printf "${RED}[FAIL]${NC} %s\n" "$*" >&2
}

# =====================================================================
# Path detection
# =====================================================================

get_script_dir() {
  # Get the directory of this script (POSIX-compliant)
  dirname "$0"
}

SCRIPT_DIR="$(cd "$(get_script_dir)" && pwd)"
KB_FILE="$SCRIPT_DIR/../vars/image_uid_gid_database.yml"

# =====================================================================
# Validation functions
# =====================================================================

is_numeric() {
  case "$1" in
  '' | *[!0-9]*) return 1 ;;
  *) return 0 ;;
  esac
}

try_detect_uid_gid() {
  image="$1"

  # Try --entrypoint sh
  result=$(docker run --rm --entrypoint sh "$image" -c 'id -u && id -g' 2>/dev/null || true)
  if [ -n "$result" ]; then
    uid=$(printf "%s" "$result" | sed -n '1p' | tr -d '\r\n')
    gid=$(printf "%s" "$result" | sed -n '2p' | tr -d '\r\n')
    if is_numeric "$uid" && is_numeric "$gid"; then
      printf "%s:%s" "$uid" "$gid"
      return 0
    fi
  fi

  # Try --entrypoint /bin/sh
  result=$(docker run --rm --entrypoint /bin/sh "$image" -c 'id -u && id -g' 2>/dev/null || true)
  if [ -n "$result" ]; then
    uid=$(printf "%s" "$result" | sed -n '1p' | tr -d '\r\n')
    gid=$(printf "%s" "$result" | sed -n '2p' | tr -d '\r\n')
    if is_numeric "$uid" && is_numeric "$gid"; then
      printf "%s:%s" "$uid" "$gid"
      return 0
    fi
  fi

  return 1
}

validate_image() {
  image="$1"
  expected_uid="$2"
  expected_gid="$3"

  # Check if image exists or can be pulled
  if ! docker image inspect "$image" >/dev/null 2>&1; then
    if ! docker pull "$image" >/dev/null 2>&1; then
      print_warning "$image - Image not found or inaccessible"
      return 2
    fi
  fi

  # Detect actual UID/GID
  if result=$(try_detect_uid_gid "$image"); then
    actual_uid=$(printf "%s" "$result" | cut -d: -f1)
    actual_gid=$(printf "%s" "$result" | cut -d: -f2)
  else
    print_warning "$image - Cannot detect UID/GID"
    return 2
  fi

  # Compare
  if [ "$actual_uid" = "$expected_uid" ] && [ "$actual_gid" = "$expected_gid" ]; then
    print_success "$image - UID: $actual_uid, GID: $actual_gid"
    return 0
  else
    print_error "$image - UID/GID mismatch"
    printf "           Expected: UID=%s, GID=%s\n" "$expected_uid" "$expected_gid"
    printf "           Actual:   UID=%s, GID=%s\n" "$actual_uid" "$actual_gid"
    return 1
  fi
}

# =====================================================================
# Main validation
# =====================================================================

main() {
  # Check if KB file exists
  if [ ! -f "$KB_FILE" ]; then
    print_error "Knowledge database not found: $KB_FILE"
    exit 1
  fi

  # Check Docker
  if ! command_exists docker; then
    print_error "Docker is not installed or not in PATH"
    exit 1
  fi

  if ! docker info >/dev/null 2>&1; then
    print_error "Docker daemon is not running"
    exit 1
  fi

  # Print header
  printf "================================================================\n"
  printf "Image UID/GID Knowledge Database Validation\n"
  printf "================================================================\n"
  printf "Database: %s\n" "$KB_FILE"
  printf "Platform: %s\n" "$OS_TYPE"
  printf "Started:  %s\n" "$(date -u 2>/dev/null || date)"
  printf "================================================================\n\n"

  # Extract and validate entries
  total=0
  passed=0
  failed=0
  skipped=0

  # Parse YAML entries (looking for exact match format)
  while IFS= read -r line; do
    # Match pattern: "image:tag": { uid: "999", gid: "999", ... }
    echo "$line" | grep -E '^\s+"[^"]+"\s*:\s*\{\s*uid:' >/dev/null 2>&1 || continue

    # Extract image name
    image=$(echo "$line" | sed -E 's/^\s+"([^"]+)".*/\1/')

    # Extract UID
    uid=$(echo "$line" | sed -E 's/.*uid:\s*"([0-9]+)".*/\1/')

    # Extract GID
    gid=$(echo "$line" | sed -E 's/.*gid:\s*"([0-9]+)".*/\1/')

    # Validate extraction
    if [ -z "$image" ] || [ -z "$uid" ] || [ -z "$gid" ]; then
      continue
    fi

    total=$((total + 1))

    if validate_image "$image" "$uid" "$gid"; then
      passed=$((passed + 1))
    else
      case $? in
      1) failed=$((failed + 1)) ;;
      2) skipped=$((skipped + 1)) ;;
      esac
    fi

    printf "\n"
  done <"$KB_FILE"

  # Print summary
  printf "================================================================\n"
  printf "Validation Summary\n"
  printf "================================================================\n"
  printf "Total entries:  %d\n" "$total"
  print_success "Passed:         $passed"
  print_error "Failed:         $failed"
  print_warning "Skipped:        $skipped"
  printf "================================================================\n"

  if [ "$failed" -gt 0 ]; then
    printf "\n"
    print_warning "Some validations failed. Please review the output above."
    exit 1
  else
    printf "\n"
    print_success "All validations passed!"
    exit 0
  fi
}

main "$@"
