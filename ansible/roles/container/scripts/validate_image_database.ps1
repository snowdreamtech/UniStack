# =====================================================================
# validate_image_database.ps1
#
# Purpose:
#   Validate all entries in the image UID/GID knowledge database.
#   PowerShell version.
#
# Compatibility:
#   - PowerShell 5.1+ (Windows)
#   - PowerShell Core 7+ (Windows/macOS/Linux)
#
# Usage:
#   .\validate_image_database.ps1
#   pwsh validate_image_database.ps1
# =====================================================================

[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# =====================================================================
# Configuration
# =====================================================================

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$KbFile = Join-Path $ScriptDir "..\vars\image_uid_gid_database.yml"

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
        "Pass" = "Green"
        "Skip" = "Yellow"
        "Fail" = "Red"
    }

    $prefix = @{
        "Info" = "[INFO]"
        "Pass" = "[PASS]"
        "Skip" = "[SKIP]"
        "Fail" = "[FAIL]"
    }

    Write-Host "$($prefix[$Level]) $Message" -ForegroundColor $colors[$Level]
}

function Test-Numeric {
    param([string]$Value)
    return $Value -match '^\d+$'
}

function Get-ActualUidGid {
    param([string]$Image)

    # Try --entrypoint sh
    try {
        $output = docker run --rm --entrypoint sh $Image -c 'id -u && id -g' 2>&1
        if ($LASTEXITCODE -eq 0 -and $output) {
            $lines = $output -split "`n"
            $uid = $lines[0].Trim()
            $gid = $lines[1].Trim()
            if ((Test-Numeric $uid) -and (Test-Numeric $gid)) {
                return @{ UID = $uid; GID = $gid }
            }
        }
    } catch {
        # Continue to next method
    }

    # Try --entrypoint /bin/sh
    try {
        $output = docker run --rm --entrypoint /bin/sh $Image -c 'id -u && id -g' 2>&1
        if ($LASTEXITCODE -eq 0 -and $output) {
            $lines = $output -split "`n"
            $uid = $lines[0].Trim()
            $gid = $lines[1].Trim()
            if ((Test-Numeric $uid) -and (Test-Numeric $gid)) {
                return @{ UID = $uid; GID = $gid }
            }
        }
    } catch {
        # Failed
    }

    return $null
}

function Test-ImageEntry {
    param(
        [string]$Image,
        [string]$ExpectedUid,
        [string]$ExpectedGid
    )

    # Check if image exists or can be pulled
    docker image inspect $Image 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        docker pull $Image 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput "$Image - Image not found or inaccessible" -Level Skip
            return 2
        }
    }

    # Detect actual UID/GID
    $actual = Get-ActualUidGid -Image $Image
    if (-not $actual) {
        Write-ColorOutput "$Image - Cannot detect UID/GID" -Level Skip
        return 2
    }

    # Compare
    if ($actual.UID -eq $ExpectedUid -and $actual.GID -eq $ExpectedGid) {
        Write-ColorOutput "$Image - UID: $($actual.UID), GID: $($actual.GID)" -Level Pass
        return 0
    } else {
        Write-ColorOutput "$Image - UID/GID mismatch" -Level Fail
        Write-Host "           Expected: UID=$ExpectedUid, GID=$ExpectedGid"
        Write-Host "           Actual:   UID=$($actual.UID), GID=$($actual.GID)"
        return 1
    }
}

# =====================================================================
# Main Execution
# =====================================================================

function Main {
    # Check if KB file exists
    if (-not (Test-Path $KbFile)) {
        Write-ColorOutput "Knowledge database not found: $KbFile" -Level Fail
        exit 1
    }

    # Check Docker
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-ColorOutput "Docker is not installed or not in PATH" -Level Fail
        exit 1
    }

    docker info 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput "Docker daemon is not running" -Level Fail
        exit 1
    }

    # Print header
    Write-Host "================================================================"
    Write-Host "Image UID/GID Knowledge Database Validation"
    Write-Host "================================================================"
    Write-Host "Database: $KbFile"
    Write-Host "Platform: $($PSVersionTable.Platform ?? 'Windows')"
    Write-Host "Started:  $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    Write-Host "================================================================"
    Write-Host ""

    # Parse YAML and validate
    $total = 0
    $passed = 0
    $failed = 0
    $skipped = 0

    $content = Get-Content $KbFile
    foreach ($line in $content) {
        # Match pattern: "image:tag": { uid: "999", gid: "999", ... }
        if ($line -match '^\s+"([^"]+)"\s*:\s*\{\s*uid:\s*"(\d+)",\s*gid:\s*"(\d+)"') {
            $image = $Matches[1]
            $uid = $Matches[2]
            $gid = $Matches[3]

            $total++

            $result = Test-ImageEntry -Image $image -ExpectedUid $uid -ExpectedGid $gid

            switch ($result) {
                0 { $passed++ }
                1 { $failed++ }
                2 { $skipped++ }
            }

            Write-Host ""
        }
    }

    # Print summary
    Write-Host "================================================================"
    Write-Host "Validation Summary"
    Write-Host "================================================================"
    Write-Host "Total entries:  $total"
    Write-ColorOutput "Passed:         $passed" -Level Pass
    Write-ColorOutput "Failed:         $failed" -Level Fail
    Write-ColorOutput "Skipped:        $skipped" -Level Skip
    Write-Host "================================================================"

    if ($failed -gt 0) {
        Write-Host ""
        Write-ColorOutput "Some validations failed. Please review the output above." -Level Skip
        exit 1
    } else {
        Write-Host ""
        Write-ColorOutput "All validations passed!" -Level Pass
        exit 0
    }
}

Main
