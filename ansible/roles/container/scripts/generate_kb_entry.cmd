@echo off
REM =====================================================================
REM generate_kb_entry.cmd
REM
REM Purpose:
REM   Generate a properly formatted knowledge database entry.
REM   Windows CMD batch script version.
REM
REM Compatibility:
REM   - Windows Command Prompt (cmd.exe)
REM
REM Usage:
REM   generate_kb_entry.cmd <image:tag> <uid> <gid> [source_url]
REM
REM Examples:
REM   generate_kb_entry.cmd mongo:7.0 999 999 "https://github.com/docker-library/mongo"
REM   generate_kb_entry.cmd postgres:16 999 999
REM =====================================================================

setlocal enabledelayedexpansion

REM Check arguments
if "%~3"=="" (
    call :show_usage
    exit /b 1
)

set "IMAGE=%~1"
set "UID=%~2"
set "GID=%~3"
set "SOURCE=%~4"

REM Validate UID is numeric
echo %UID% | findstr /R "^[0-9][0-9]*$" >nul
if errorlevel 1 (
    echo [ERROR] UID must be numeric: %UID%
    exit /b 1
)

REM Validate GID is numeric
echo %GID% | findstr /R "^[0-9][0-9]*$" >nul
if errorlevel 1 (
    echo [ERROR] GID must be numeric: %GID%
    exit /b 1
)

REM Auto-detect source if not provided
if "%SOURCE%"=="" (
    echo %IMAGE% | findstr /C:"/" >nul
    if errorlevel 1 (
        REM Official image
        for /f "tokens=1 delims=:" %%a in ("%IMAGE%") do set "img_name=%%a"
        set "SOURCE=https://github.com/docker-library/!img_name!"
    ) else (
        REM Image with registry
        for /f "tokens=1 delims=:" %%a in ("%IMAGE%") do set "img_without_tag=%%a"
        set "SOURCE=https://hub.docker.com/r/!img_without_tag!"
    )
    echo [WARN] Auto-detected source: !SOURCE!
)

REM Generate exact match entry
echo.
echo # -------------------------------------------------------------
echo # Exact Match Entry (Recommended)
echo # -------------------------------------------------------------
echo   "%IMAGE%": { uid: "%UID%", gid: "%GID%", source: "%SOURCE%" }

REM Generate pattern match suggestion if applicable
for /f "tokens=2 delims=:" %%a in ("%IMAGE%") do set "tag=%%a"
echo !tag! | findstr /R "^[0-9]" >nul
if not errorlevel 1 (
    echo.
    echo # -------------------------------------------------------------
    echo # Pattern Match Entry (Optional - for version series)
    echo # -------------------------------------------------------------
    for /f "tokens=1 delims=:" %%b in ("%IMAGE%") do set "base=%%b"

    echo !tag! | findstr /R "^[0-9][0-9]*\.[0-9]" >nul
    if not errorlevel 1 (
        echo   - regex: "^!base!:[0-9]+\\.?[0-9]*"
    ) else (
        echo   - regex: "^!base!:[0-9]+"
    )
    echo     uid: "%UID%"
    echo     gid: "%GID%"
    echo     description: "!base! official images"
)

echo.
echo [OK] Entry generated successfully!
echo.
echo [INFO] Next steps:
echo   1. Review the generated entry
echo   2. Copy to roles\container\vars\image_uid_gid_database.yml
echo   3. Run: validate_image_database.cmd
echo   4. Commit changes to version control
echo.

exit /b 0

:show_usage
echo Usage: generate_kb_entry.cmd ^<image:tag^> ^<uid^> ^<gid^> [source_url]
echo.
echo Generate a YAML entry for the image UID/GID knowledge database.
echo.
echo Arguments:
echo   image:tag    Container image with tag (e.g., mongo:7.0)
echo   uid          User ID (numeric)
echo   gid          Group ID (numeric)
echo   source_url   Optional source URL (auto-detected if omitted)
echo.
echo Examples:
echo   generate_kb_entry.cmd mongo:7.0 999 999 "https://github.com/docker-library/mongo"
echo   generate_kb_entry.cmd postgres:16 999 999
echo   generate_kb_entry.cmd mycompany/app:latest 1001 1001
echo.
exit /b 0
