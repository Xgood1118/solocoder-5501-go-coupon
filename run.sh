#!/bin/bash
echo "========================================"
echo "Coupon Service - Startup Script (Linux/Mac)"
echo "========================================"

if ! command -v go &> /dev/null; then
    echo "[ERROR] Go is not installed. Please install Go first."
    echo "Download from: https://golang.org/dl/"
    exit 1
fi

echo "[INFO] Go found:"
go version

echo ""
echo "[INFO] Downloading dependencies..."
go mod tidy || { echo "[ERROR] Failed to download dependencies"; exit 1; }

echo ""
echo "[INFO] Building..."
go build -o coupon-service . || { echo "[ERROR] Build failed"; exit 1; }

echo ""
echo "[INFO] Starting coupon service..."
echo "  Port: 8080 (default)"
echo "  DB: ./coupon.db (default)"
echo "  Ticker: 1 minute (default)"
echo ""
echo "Environment variables (optional):"
echo "  export COUPON_PORT=8080"
echo "  export COUPON_DB_PATH=./coupon.db"
echo "  export COUPON_TICKER_MINUTES=1"
echo ""
echo "Press Ctrl+C to stop."
echo "========================================"

./coupon-service
