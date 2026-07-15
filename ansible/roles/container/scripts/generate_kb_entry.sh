#!/bin/sh
# =====================================================================
# generate_kb_entry.sh
#
# Purpose:
#   Generate a properly formatted knowledge database entry.
#
# Compatibility:
#   - POSIX sh compliant (no bash-specific features)
#   - macOS, Linux, Windows (Git Bash/WSL/MSYS2)
#
# Usage:
#   ./generate_kb_entry.sh <image:tag> <uid> <gid> [source_url]
#
# Examples:
#   ./generate_kb_entry.sh mongo:7.0 999 999 "https://github.com/docker-library/mongo"
#   ./generate_kb_entry.sh postgres:16 999 999
# =====================================================================

set -eu

# =====================================================================
# Cross-platform compatibility
# =====================================================================

supports_color() {
  if [ -t 1 ]; then
    if [ -n "${TERM:-}" ] && [ "$TERM" != "dumb" ]; then
      return 0
    fi
  fi
  return 1
}

if supports_color; then
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  RED='\033[0;31m'
  BLUE='\033[0;34m'
  NC='\033[0m'
else
  GREEN=''
  YELLOW=''
  RED=''
  BLUE=''
  NC=''
fi

print_success() {
  printf "${GREEN}[OK]${NC} %s\n" "$*"
}

print_warning() {
  printf "${YELLOW}[WARN]${NC} %s\n" "$*"
}

print_error() {
  printf "${RED}[ERROR]${NC} %s\n" "$*" >&2
}

print_info() {
  printf "${BLUE}[INFO]${NC} %s\n" "$*"
}

# =====================================================================
# Validation
# =====================================================================

is_numeric() {
  case "$1" in
  '' | *[!0-9]*) return 1 ;;
  *) return 0 ;;
  esac
}

has_registry() {
  case "$1" in
  */*) return 0 ;;
  *) return 1 ;;
  esac
}

get_image_name() {
  printf "%s" "$1" | sed 's/:.*$//'
}

get_image_tag() {
  printf "%s" "$1" | sed 's/^.*://'
}

# =====================================================================
# Source URL generation
# =====================================================================

generate_source_url() {
  image="$1"

  if has_registry "$image"; then
    image_without_tag=$(get_image_name "$image")
    printf "https://hub.docker.com/r/%s" "$image_without_tag"
  else
    image_name=$(get_image_name "$image")
    printf "https://github.com/docker-library/%s" "$image_name"
  fi
}

# =====================================================================
# Pattern generation
# =====================================================================

generate_pattern_entry() {
  image="$1"
  uid="$2"
  gid="$3"

  image_base=$(get_image_name "$image")
  image_tag=$(get_image_tag "$image")

  # Check if tag looks like a version number
  case "$image_tag" in
  [0-9]*.[0-9]*)
    # Looks like a version (e.g., 7.0, 16.2)
    cat <<EOF

# ─────────────────────────────────────────────────────────────
# Pattern Match Entry (Optional - for version series)
# ─────────────────────────────────────────────────────────────
  - regex: "^${image_base}:[0-9]+\\\\.?[0-9]*"
    uid: "${uid}"
    gid: "${gid}"
    description: "${image_base} official images"
EOF
    ;;
  [0-9]*)
    # Looks like a major version (e.g., 7, 16)
    cat <<EOF

# ─────────────────────────────────────────────────────────────
# Pattern Match Entry (Optional - for version series)
# ─────────────────────────────────────────────────────────────
  - regex: "^${image_base}:[0-9]+"
    uid: "${uid}"
    gid: "${gid}"
    description: "${image_base} official images"
EOF
    ;;
  esac
}

# =====================================================================
# Usage
# =====================================================================

show_usage() {
  cat <<'EOF'
Usage: generate_kb_entry.sh <image:tag> <uid> <gid> [source_url]

Generate a YAML entry for the image UID/GID knowledge database.

Arguments:
  image:tag    Container image with tag (e.g., mongo:7.0)
  uid          User ID (numeric)
  gid          Group ID (numeric)
  source_url   Optional source URL (auto-detected if omitted)

Examples:
  generate_kb_entry.sh mongo:7.0 999 999 "https://github.com/docker-library/mongo"
  generate_kb_entry.sh postgres:16 999 999
  generate_kb_entry.sh mycompany/app:latest 1001 1001 "https://github.com/mycompany/app"

Output:
  YAML-formatted entries ready to copy to:
  roles/container/vars/image_uid_gid_database.yml

Compatibility:
  - macOS, Linux, Windows (Git Bash/WSL/MSYS2)
  - POSIX sh compliant
EOF
}

# =====================================================================
# Main
# =====================================================================

main() {
  # Validate arguments
  if [ $# -lt 3 ]; then
    show_usage
    exit 1
  fi

  IMAGE="$1"
  UID="$2"
  GID="$3"
  SOURCE="${4:-}"

  # Validate UID/GID are numeric
  if ! is_numeric "$UID"; then
    print_error "UID must be numeric: $UID"
    exit 1
  fi

  if ! is_numeric "$GID"; then
    print_error "GID must be numeric: $GID"
    exit 1
  fi

  # Auto-detect source if not provided
  if [ -z "$SOURCE" ]; then
    SOURCE=$(generate_source_url "$IMAGE")
    print_warning "Auto-detected source: $SOURCE"
  fi

  # Generate exact match entry
  printf "\n"
  printf "# ─────────────────────────────────────────────────────────────\n"
  printf "# Exact Match Entry (Recommended)\n"
  printf "# ─────────────────────────────────────────────────────────────\n"
  printf "  \"%s\": { uid: \"%s\", gid: \"%s\", source: \"%s\" }\n" "$IMAGE" "$UID" "$GID" "$SOURCE"

  # Generate pattern match suggestion if applicable
  generate_pattern_entry "$IMAGE" "$UID" "$GID"

  printf "\n"
  print_success "Entry generated successfully!"
  printf "\n"
  print_info "Next steps:"
  printf "  1. Review the generated entry\n"
  printf "  2. Copy to roles/container/vars/image_uid_gid_database.yml\n"
  printf "  3. Run: ./validate_image_database.sh\n"
  printf "  4. Commit changes to version control\n"
  printf "\n"
}

main "$@"
