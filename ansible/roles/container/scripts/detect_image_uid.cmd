@echo off
REM =====================================================================
REM detect_image_uid.cmd
REM
REM Purpose:
REM   Automatically detect UID/GID for container images.
REM   Windows CMD batch script version.
REM
REM Compatibility:
REM   - Windows Command Prompt (cmd.exe)
REM   - Requires Docker Desktop for Windows
REM
REM Usage:
REM   detect_image_uid.cmd <image:tag> [image:tag ...]
REM
REM Examples:
REM   detect_image_uid.cmd mongo:7.0
REM   detect_image_uid.cmd mongo:7.0 postgres:16
REM =====================================================================

setlocal enabledelayedexpansion

REM Check if arguments provided
if "%~1"=="" (
    call :show_usage
    exit /b 1
)

REM Check Docker availability
call :check_docker
if errorlevel 1 exit /b 1

REM Print header
echo # ==============================================
echo # Image UID/GID Detection Results
echo # Generated: %date% %time%
echo # Platform: Windows (CMD)
echo # ==============================================
echo.
echo # Add these entries to image_uid_gid_exact_match:
echo.

set success_count=0
set fail_count=0

REM Process each image
:process_loop
if "%~1"=="" goto :summary

set "image=%~1"
call :detect_image_uid_gid "%image%"
if errorlevel 1 (
    set /a fail_count+=1
) else (
    set /a success_count+=1
)
echo.

shift
goto :process_loop

REM Print summary
:summary
echo.
echo # ==============================================
echo # Summary: %success_count% succeeded, %fail_count% failed
echo # ==============================================

if %fail_count% gtr 0 exit /b 1
exit /b 0

REM =====================================================================
REM Functions
REM =====================================================================

:check_docker
where docker >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed or not in PATH
    echo [INFO] Please install Docker Desktop: https://docs.docker.com/desktop/install/windows-install/
    exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker daemon is not running
    echo [INFO] Please start Docker Desktop
    exit /b 1
)
exit /b 0

:detect_image_uid_gid
set "img=%~1"
echo [INFO] Detecting UID/GID for: %img%

REM Pull image if needed
docker image inspect "%img%" >nul 2>&1
if errorlevel 1 (
    echo [INFO] Pulling image: %img%
    docker pull "%img%" >nul 2>&1
    if errorlevel 1 (
        echo [ERROR] Failed to pull image: %img%
        exit /b 1
    )
)

REM Try --entrypoint sh
for /f "tokens=*" %%a in ('docker run --rm --entrypoint sh "%img%" -c "id -u && id -g" 2^>nul') do (
    if not defined uid (
        set "uid=%%a"
    ) else if not defined gid (
        set "gid=%%a"
    )
)

REM Try --entrypoint /bin/sh if first method failed
if not defined uid (
    for /f "tokens=*" %%a in ('docker run --rm --entrypoint /bin/sh "%img%" -c "id -u && id -g" 2^>nul') do (
        if not defined uid (
            set "uid=%%a"
        ) else if not defined gid (
            set "gid=%%a"
        )
    )
)

REM Try docker inspect if still not found
if not defined uid (
    for /f "tokens=*" %%a in ('docker inspect -f "{{.Config.User}}" "%img%" 2^>nul') do (
        set "user_info=%%a"
    )
    if defined user_info (
        for /f "tokens=1,2 delims=:" %%a in ("!user_info!") do (
            set "uid=%%a"
            set "gid=%%b"
            if not defined gid set "gid=%%a"
        )
    )
)

if not defined uid (
    echo [ERROR] Failed to detect UID/GID for: %img%
    exit /b 1
)

REM Generate source URL
set "source_url="
echo %img% | findstr /C:"/" >nul
if errorlevel 1 (
    REM Official image
    for /f "tokens=1 delims=:" %%a in ("%img%") do set "img_name=%%a"
    set "source_url=https://github.com/docker-library/!img_name!"
) else (
    REM Image with registry
    for /f "tokens=1 delims=:" %%a in ("%img%") do set "img_without_tag=%%a"
    set "source_url=https://hub.docker.com/r/!img_without_tag!"
)

REM Output YAML format
echo.
echo   "%img%": { uid: "%uid%", gid: "%gid%", source: "%source_url%" }

echo [OK] Generated YAML entry for: %img%

REM Clean up variables
set "uid="
set "gid="
set "user_info="
set "source_url="

exit /b 0

:show_usage
echo Usage: detect_image_uid.cmd ^<image:tag^> [image:tag ...]
echo.
echo Detects UID/GID for container images and outputs YAML format.
echo.
echo Arguments:
echo   image:tag     One or more container images with tags
echo.
echo Examples:
echo   detect_image_uid.cmd mongo:7.0
echo   detect_image_uid.cmd mongo:7.0 postgres:16
echo.
echo Requirements:
echo   - Docker Desktop for Windows must be installed and running
echo   - Network access to pull images (if not cached)
echo.
exit /b 0
