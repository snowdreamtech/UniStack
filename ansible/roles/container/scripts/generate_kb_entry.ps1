# =====================================================================
# generate_kb_entry.ps1
#
# Purpose:
#   Generate a properly formatted knowledge database entry.
#   PowerShell version.
#
# Compatibility:
#   - PowerShell 5.1+ (Windows)
#   - PowerShell Core 7+ (Windows/macOS/Linux)
#
# Usage:
#   .\generate_kb_entry.ps1 <image:tag> <uid> <gid> [source_url]
#   pwsh generate_kb_entry.ps1 <image:tag> <uid> <gid> [source_url]
#
# Examples:
#   .\generate_kb_entry.ps1 mongo:7.0 999 999 "https://github.com/docker-library/mongo"
#   .\generate_kb_entry.ps1 postgres:16 999 999
# =====================================================================

[CmdletBinding()]
param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Image,

    [Parameter(Mandatory=$true, Position=1)]
    [string]$UID,

    [Parameter(Mandatory=$true, Position=2)]
    [string]$GID,

    [Parameter(Mandatory=$false, Position=3)]
    [string]$Source = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# =====================================================================
# Helper Functions
# =====================================================================

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Level = "Info"
    )

    $colors = @{
        "Info" = "Cyan"
        "Success" = "Green"
        "Warning" = "Yellow"
        "Error" = "Red"
    }

    $prefix = @{
        "Info" = "[INFO]"
        "Success" = "[OK]"
        "Warning" = "[WARN]"
        "Error" = "[ERROR]"
    }

    Write-Host "$($prefix[$Level]) $Message" -ForegroundColor $colors[$Level]
}

function Test-Numeric {
    param([string]$Value)
    return $Value -match '^\d+$'
}

function Get-SourceUrl {
    param([string]$Image)

    if ($Image -match '/') {
        $imageWithoutTag = $Image -replace ':.*$', ''
        return "https://hub.docker.com/r/$imageWithoutTag"
    } else {
        $imageName = $Image -replace ':.*$', ''
        return "https://github.com/docker-library/$imageName"
    }
}

function Get-PatternEntry {
    param(
        [string]$Image,
        [string]$UID,
        [string]$GID
    )

    $parts = $Image -split ':'
    $base = $parts[0]
    $tag = $parts[1]

    if ($tag -match '^\d+\.\d+') {
        # Version like 7.0
        return @"

# -------------------------------------------------------------
# Pattern Match Entry (Optional - for version series)
# -------------------------------------------------------------
  - regex: "^$base:[0-9]+\\.?[0-9]*"
    uid: "$UID"
    gid: "$GID"
    description: "$base official images"
"@
    } elseif ($tag -match '^\d+') {
        # Version like 7
        return @"

# -------------------------------------------------------------
# Pattern Match Entry (Optional - for version series)
# -------------------------------------------------------------
  - regex: "^$base:[0-9]+"
    uid: "$UID"
    gid: "$GID"
    description: "$base official images"
"@
    }

    return ""
}

function Show-Usage {
    Write-Host @"
Usage: generate_kb_entry.ps1 <image:tag> <uid> <gid> [source_url]

Generate a YAML entry for the image UID/GID knowledge database.

Arguments:
  image:tag    Container image with tag (e.g., mongo:7.0)
  uid          User ID (numeric)
  gid          Group ID (numeric)
  source_url   Optional source URL (auto-detected if omitted)

Examples:
  .\generate_kb_entry.ps1 mongo:7.0 999 999 "https://github.com/docker-library/mongo"
  .\generate_kb_entry.ps1 postgres:16 999 999
  .\generate_kb_entry.ps1 mycompany/app:latest 1001 1001

Output:
  YAML-formatted entries ready to copy to:
  roles/container/vars/image_uid_gid_database.yml

Compatibility:
  - Windows PowerShell 5.1+
  - PowerShell Core 7+ (Windows/macOS/Linux)
"@
}

# =====================================================================
# Main Execution
# =====================================================================

function Main {
    # Validate UID is numeric
    if (-not (Test-Numeric $UID)) {
        Write-ColorOutput "UID must be numeric: $UID" -Level Error
        exit 1
    }

    # Validate GID is numeric
    if (-not (Test-Numeric $GID)) {
        Write-ColorOutput "GID must be numeric: $GID" -Level Error
        exit 1
    }

    # Auto-detect source if not provided
    if ([string]::IsNullOrEmpty($Source)) {
        $Source = Get-SourceUrl -Image $Image
        Write-ColorOutput "Auto-detected source: $Source" -Level Warning
    }

    # Generate exact match entry
    Write-Host ""
    Write-Host "# -------------------------------------------------------------"
    Write-Host "# Exact Match Entry (Recommended)"
    Write-Host "# -------------------------------------------------------------"
    Write-Host "  `"$Image`": { uid: `"$UID`", gid: `"$GID`", source: `"$Source`" }"

    # Generate pattern match suggestion if applicable
    $patternEntry = Get-PatternEntry -Image $Image -UID $UID -GID $GID
    if ($patternEntry) {
        Write-Host $patternEntry
    }

    Write-Host ""
    Write-ColorOutput "Entry generated successfully!" -Level Success
    Write-Host ""
    Write-ColorOutput "Next steps:" -Level Info
    Write-Host "  1. Review the generated entry"
    Write-Host "  2. Copy to roles/container/vars/image_uid_gid_database.yml"
    Write-Host "  3. Run: .\validate_image_database.ps1"
    Write-Host "  4. Commit changes to version control"
    Write-Host ""

    exit 0
}

Main
