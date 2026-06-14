@echo off
echo ========================================
echo Coupon Service - Startup Script (Windows)
echo ========================================

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed. Please install Go first.
    echo Download from: https://golang.org/dl/
    echo After installation, restart this script.
    pause
    exit /b 1
)

echo [INFO] Go found:
go version

set CGO_ENABLED=0
echo [INFO] CGO_ENABLED=%CGO_ENABLED% (pure Go build, no gcc required)

if exist coupon.db (
    echo [WARN] Old coupon.db found, deleting to avoid compatibility issues...
    del /F /Q coupon.db 2>nul
    echo [INFO] Old database deleted.
)
if exist coupon.db-wal del /F /Q coupon.db-wal 2>nul
if exist coupon.db-shm del /F /Q coupon.db-shm 2>nul

echo.
echo [INFO] Downloading dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo [ERROR] Failed to download dependencies.
    pause
    exit /b 1
)

echo.
echo [INFO] Building pure Go binary (no cgo)...
go build -tags=sqlite_omit_load_extension -ldflags="-s -w" -o coupon-service.exe .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    pause
    exit /b 1
)

echo.
echo [INFO] Build successful! Binary size:
for %%I in (coupon-service.exe) do echo   %%~zI bytes

echo.
echo [INFO] Starting coupon service...
echo   Port: 8080 (default)
echo   DB: ./coupon.db (default)
echo   Ticker: 1 minute (default)
echo.
echo Environment variables (optional):
echo   set COUPON_PORT=8080
echo   set COUPON_DB_PATH=./coupon.db
echo   set COUPON_TICKER_MINUTES=1
echo.
echo Press Ctrl+C to stop.
echo ========================================

coupon-service.exe
pause
