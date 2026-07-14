@echo off
REM =====================================================================
REM validate_image_database.cmd
REM
REM Purpose:
REM   Validate all entries in the image UID/GID knowledge database.
REM   Windows CMD batch script version.
REM
REM Compatibility:
REM   - Windows Command Prompt (cmd.exe)
REM   - Requires Docker Desktop for Windows
REM
REM Usage:
REM   validate_image_database.cmd
REM =====================================================================

setlocal enabledelayedexpansion

REM Get script directory
set "SCRIPT_DIR=%~dp0"
set "KB_FILE=%SCRIPT_DIR%..\vars\image_uid_gid_database.yml"

REM Check if KB file exists
if not exist "%KB_FILE%" (
    echo [ERROR] Knowledge database not found: %KB_FILE%
    exit /b 1
)

REM Check Docker
where docker >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker daemon is not running
    exit /b 1
)

REM Print header
echo ================================================================
echo Image UID/GID Knowledge Database Validation
echo ================================================================
echo Database: %KB_FILE%
echo Platform: Windows (CMD)
echo Started:  %date% %time%
echo ================================================================
echo.

set total=0
set passed=0
set failed=0
set skipped=0

REM Parse YAML and validate entries
for /f "usebackq tokens=*" %%a in ("%KB_FILE%") do (
    set "line=%%a"
    echo !line! | findstr /R /C:"^  \".*\":.*{ *uid:" >nul
    if not errorlevel 1 (
        REM Extract image name
        for /f "tokens=2 delims=`"" %%b in ("!line!") do set "image=%%b"

        REM Extract UID
        for /f "tokens=2 delims=`"" %%c in ("!line!") do (
            echo %%c | findstr /C:"uid:" >nul
            if not errorlevel 1 (
                for /f "tokens=2 delims=:" %%d in ("%%c") do set "uid=%%d"
            )
        )

        REM Extract GID (simplified - assumes format)
        for /f "tokens=4 delims=`"" %%e in ("!line!") do set "gid=%%e"

        if defined image if defined uid if defined gid (
            set /a total+=1
            call :validate_image "!image!" "!uid!" "!gid!"

            set "image="
            set "uid="
            set "gid="
        )
    )
)

REM Print summary
echo.
echo ================================================================
echo Validation Summary
echo ================================================================
echo Total entries:  %total%
echo [PASS] Passed:  %passed%
echo [FAIL] Failed:  %failed%
echo [SKIP] Skipped: %skipped%
echo ================================================================

if %failed% gtr 0 (
    echo.
    echo [WARN] Some validations failed. Please review the output above.
    exit /b 1
) else (
    echo.
    echo [PASS] All validations passed!
    exit /b 0
)

REM =====================================================================
REM Functions
REM =====================================================================

:validate_image
set "img=%~1"
set "exp_uid=%~2"
set "exp_gid=%~3"

REM Check if image exists or can be pulled
docker image inspect "%img%" >nul 2>&1
if errorlevel 1 (
    docker pull "%img%" >nul 2>&1
    if errorlevel 1 (
        echo [SKIP] %img% - Image not found or inaccessible
        set /a skipped+=1
        exit /b 0
    )
)

REM Detect actual UID/GID
set "act_uid="
set "act_gid="

for /f "tokens=*" %%a in ('docker run --rm --entrypoint sh "%img%" -c "id -u && id -g" 2^>nul') do (
    if not defined act_uid (
        set "act_uid=%%a"
    ) else if not defined act_gid (
        set "act_gid=%%a"
    )
)

if not defined act_uid (
    echo [SKIP] %img% - Cannot detect UID/GID
    set /a skipped+=1
    exit /b 0
)

REM Compare
if "%act_uid%"=="%exp_uid%" if "%act_gid%"=="%exp_gid%" (
    echo [PASS] %img% - UID: %act_uid%, GID: %act_gid%
    set /a passed+=1
) else (
    echo [FAIL] %img% - UID/GID mismatch
    echo            Expected: UID=%exp_uid%, GID=%exp_gid%
    echo            Actual:   UID=%act_uid%, GID=%act_gid%
    set /a failed+=1
)

exit /b 0
