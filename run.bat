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

echo.
echo [INFO] Downloading dependencies...
go mod tidy
if %errorlevel% neq 0 (
    echo [ERROR] Failed to download dependencies.
    pause
    exit /b 1
)

echo.
echo [INFO] Building...
go build -o coupon-service.exe .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed.
    pause
    exit /b 1
)

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
