# =====================================================================
# detect_image_uid.ps1
#
# Purpose:
#   Automatically detect UID/GID for container images.
#   PowerShell version for Windows.
#
# Compatibility:
#   - PowerShell 5.1+ (Windows)
#   - PowerShell Core 7+ (Windows/macOS/Linux)
#
# Usage:
#   .\detect_image_uid.ps1 <image:tag> [image:tag ...]
#   pwsh detect_image_uid.ps1 <image:tag> [image:tag ...]
#
# Examples:
#   .\detect_image_uid.ps1 mongo:7.0
#   .\detect_image_uid.ps1 mongo:7.0 postgres:16 redis:7-alpine
# =====================================================================

[CmdletBinding()]
param(
    [Parameter(Mandatory=$false, ValueFromRemainingArguments=$true)]
    [string[]]$Images
)

# Strict mode
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

    if ($Host.UI.SupportsVirtualTerminal -or $PSVersionTable.PSVersion.Major -ge 6) {
        Write-Host "$($prefix[$Level]) $Message" -ForegroundColor $colors[$Level]
    } else {
        Write-Host "$($prefix[$Level]) $Message"
    }
}

function Test-DockerAvailable {
    # Check if Docker command exists
    $dockerCmd = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $dockerCmd) {
        Write-ColorOutput "Docker is not installed or not in PATH" -Level Error
        Write-ColorOutput "Please install Docker Desktop: https://docs.docker.com/get-docker/" -Level Info
        return $false
    }

    # Test if Docker daemon is running
    try {
        docker info 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw
        }
    } catch {
        Write-ColorOutput "Docker daemon is not running" -Level Error
        Write-ColorOutput "Please start Docker Desktop or Docker service" -Level Info
        return $false
    }

    return $true
}

function Test-Numeric {
    param([string]$Value)
    return $Value -match '^\d+$'
}

function Get-ImageUidGid {
    param(
        [string]$Image
    )

    Write-ColorOutput "Detecting UID/GID for: $Image" -Level Info

    # Pull image if needed
    try {
        docker image inspect $Image 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput "Pulling image: $Image" -Level Info
            docker pull $Image 2>&1 | Out-Null
            if ($LASTEXITCODE -ne 0) {
                throw "Failed to pull image"
            }
        }
    } catch {
        Write-ColorOutput "Failed to pull image: $Image" -Level Error
        return $null
    }

    $uid = $null
    $gid = $null
    $method = $null

    # Method 1: Try --entrypoint sh
    try {
        $output = docker run --rm --entrypoint sh $Image -c 'id -u && id -g' 2>&1
        if ($LASTEXITCODE -eq 0 -and $output) {
            $lines = $output -split "`n"
            $uid = $lines[0].Trim()
            $gid = $lines[1].Trim()
            if ((Test-Numeric $uid) -and (Test-Numeric $gid)) {
                $method = "--entrypoint sh"
                Write-ColorOutput "Detected via --entrypoint sh: UID=$uid, GID=$gid" -Level Success
            } else {
                $uid = $null
                $gid = $null
            }
        }
    } catch {
        # Silently continue to next method
    }

    # Method 2: Try --entrypoint /bin/sh
    if (-not $uid) {
        try {
            $output = docker run --rm --entrypoint /bin/sh $Image -c 'id -u && id -g' 2>&1
            if ($LASTEXITCODE -eq 0 -and $output) {
                $lines = $output -split "`n"
                $uid = $lines[0].Trim()
                $gid = $lines[1].Trim()
                if ((Test-Numeric $uid) -and (Test-Numeric $gid)) {
                    $method = "--entrypoint /bin/sh"
                    Write-ColorOutput "Detected via --entrypoint /bin/sh: UID=$uid, GID=$gid" -Level Warning
                } else {
                    $uid = $null
                    $gid = $null
                }
            }
        } catch {
            # Silently continue
        }
    }

    # Method 3: Try docker inspect
    if (-not $uid) {
        try {
            $userInfo = docker inspect -f '{{.Config.User}}' $Image 2>&1
            if ($LASTEXITCODE -eq 0 -and $userInfo) {
                $userInfo = $userInfo.Trim()
                if ($userInfo -match '^(\d+):(\d+)$') {
                    $uid = $Matches[1]
                    $gid = $Matches[2]
                    $method = "docker inspect"
                    Write-ColorOutput "Detected via docker inspect: UID=$uid, GID=$gid" -Level Warning
                } elseif ($userInfo -match '^\d+$') {
                    $uid = $userInfo
                    $gid = $userInfo
                    $method = "docker inspect"
                    Write-ColorOutput "Detected via docker inspect: UID=$uid, GID=$gid" -Level Warning
                }
            }
        } catch {
            # Silently continue
        }
    }

    if (-not $uid -or -not $gid) {
        Write-ColorOutput "Failed to detect UID/GID for: $Image" -Level Error
        return $null
    }

    # Generate source URL
    $sourceUrl = if ($Image -match '/') {
        $imageWithoutTag = $Image -replace ':.*$', ''
        "https://hub.docker.com/r/$imageWithoutTag"
    } else {
        $imageName = $Image -replace ':.*$', ''
        "https://github.com/docker-library/$imageName"
    }

    return @{
        Image = $Image
        UID = $uid
        GID = $gid
        Source = $sourceUrl
        Method = $method
    }
}

function Show-Usage {
    Write-Host @"
Usage: detect_image_uid.ps1 <image:tag> [image:tag ...]

Detects UID/GID for container images and outputs YAML format.

Arguments:
  image:tag     One or more container images with tags

Examples:
  .\detect_image_uid.ps1 mongo:7.0
  .\detect_image_uid.ps1 mongo:7.0 postgres:16 redis:7-alpine

Output:
  YAML-formatted entries ready to copy to:
  roles/container/vars/image_uid_gid_database.yml

Requirements:
  - Docker must be installed and running
  - PowerShell 5.1+ or PowerShell Core 7+
  - Network access to pull images (if not cached)

Compatibility:
  - Windows PowerShell 5.1+
  - PowerShell Core 7+ (Windows/macOS/Linux)
"@
}

# =====================================================================
# Main Execution
# =====================================================================

function Main {
    # Check arguments
    if (-not $Images -or $Images.Count -eq 0) {
        Show-Usage
        exit 1
    }

    # Check Docker availability
    if (-not (Test-DockerAvailable)) {
        exit 1
    }

    # Print header
    Write-Host "# =============================================="
    Write-Host "# Image UID/GID Detection Results"
    Write-Host "# Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC')"
    Write-Host "# Platform: $($PSVersionTable.Platform ?? 'Windows')"
    Write-Host "# PowerShell: $($PSVersionTable.PSVersion)"
    Write-Host "# =============================================="
    Write-Host ""
    Write-Host "# Add these entries to image_uid_gid_exact_match:"
    Write-Host ""

    # Process each image
    $successCount = 0
    $failCount = 0

    foreach ($image in $Images) {
        $result = Get-ImageUidGid -Image $image

        if ($result) {
            Write-Host ""
            Write-Host "  `"$($result.Image)`": { uid: `"$($result.UID)`", gid: `"$($result.GID)`", source: `"$($result.Source)`" }"
            Write-ColorOutput "Generated YAML entry for: $($result.Image) (method: $($result.Method))" -Level Success
            $successCount++
        } else {
            $failCount++
        }

        Write-Host ""
    }

    # Print summary
    Write-Host ""
    Write-Host "# =============================================="
    Write-Host "# Summary: $successCount succeeded, $failCount failed"
    Write-Host "# =============================================="

    if ($failCount -gt 0) {
        exit 1
    }
    exit 0
}

# Run main function
Main
