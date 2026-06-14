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

export CGO_ENABLED=0
echo "[INFO] CGO_ENABLED=$CGO_ENABLED (pure Go build, no gcc required)"

if [ -f coupon.db ]; then
    echo "[WARN] Old coupon.db found, deleting to avoid compatibility issues..."
    rm -f coupon.db coupon.db-wal coupon.db-shm
    echo "[INFO] Old database deleted."
fi

echo ""
echo "[INFO] Downloading dependencies..."
go mod tidy || { echo "[ERROR] Failed to download dependencies"; exit 1; }

echo ""
echo "[INFO] Building pure Go binary (no cgo)..."
go build -tags=sqlite_omit_load_extension -ldflags="-s -w" -o coupon-service . || { echo "[ERROR] Build failed"; exit 1; }

echo ""
echo "[INFO] Build successful! Binary size: $(ls -lh coupon-service | awk '{print $5}')"

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
