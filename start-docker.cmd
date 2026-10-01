echo off
goto(){
    # Windows jumps to the label below; POSIX shells execute this function.
    case "${1-}" in
        '') loop_mode=build ;;
        --no-build) loop_mode=start ;;
        --help|-h) printf '%s\n' 'Usage: sh start-docker.cmd [--no-build]'; return 0 ;;
        *) printf '%s\n' 'Unknown option. Use --no-build to start existing images.' >&2; return 1 ;;
    esac
    [ "$#" -le 1 ] || { printf '%s\n' 'Too many arguments.' >&2; return 1; }
    command -v docker >/dev/null 2>&1 || { printf '%s\n' 'Install Docker with the Compose plugin first.' >&2; return 1; }
    docker info >/dev/null 2>&1 || { printf '%s\n' 'Start Docker Desktop or the Docker daemon, then retry.' >&2; return 1; }
    docker compose version || return 1
    loop_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd) || return 1
    export API_PORT=8081
    if [ "$loop_mode" = build ]; then
        docker compose --project-directory "$loop_root" -f "$loop_root/docker-compose.yml" up --build -d --wait --wait-timeout 90 || return 1
    else
        docker compose --project-directory "$loop_root" -f "$loop_root/docker-compose.yml" up --no-build --pull never -d --wait --wait-timeout 90 || return 1
    fi
    printf '%s\n' 'Loop containers are running: http://localhost:3000' 'API: http://localhost:8081' 'To stop, run docker compose down in the project folder.'
}
goto "$@"
exit $?
:(){
@echo off
setlocal
if "%~1"=="--help" goto help
if "%~1"=="-h" goto help
if not "%~2"=="" goto usage_error
set "LOOP_OPTIONS=--build"
if "%~1"=="--no-build" (
    set "LOOP_OPTIONS=--no-build --pull never"
) else if not "%~1"=="" goto usage_error
where docker >nul 2>&1
if errorlevel 1 (
    echo Install Docker with the Compose plugin first.
    exit /b 1
)
docker info >nul 2>&1
if errorlevel 1 (
    echo Start Docker Desktop and wait for its engine, then retry.
    exit /b 1
)
docker compose version
if errorlevel 1 exit /b 1
set "API_PORT=8081"
docker compose --project-directory "%~dp0." -f "%~dp0docker-compose.yml" up %LOOP_OPTIONS% -d --wait --wait-timeout 90
if errorlevel 1 (
    echo Startup failed. Check the output above. If images are missing, run without --no-build.
    exit /b 1
)
echo Loop containers are running: http://localhost:3000
echo API: http://localhost:8081
echo To stop, run docker compose down in the project folder.
exit /b 0
:usage_error
echo Unknown option or too many arguments. Use --no-build to start existing images.
exit /b 1
:help
echo Usage: start-docker.cmd [--no-build]
exit /b 0
