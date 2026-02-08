#!/bin/bash

echo "========================================"
echo "   SmartBell Build for Armbian (ARM) 🐧"
echo "========================================"
echo ""

echo "[*] Compiling binary for ARM64..."

# Extract version from main.go (e.g., "v1.1.0" -> "v1_1_0")
VERSION=$(grep 'AppVersion.*=' main.go | grep -oP 'v[0-9]+\.[0-9]+\.[0-9]+' | tr '.' '_')
DATE=$(date +"%d%m%y")
OUT_FILE="smartbell_${VERSION}_${DATE}_arm64"

echo "[*] Version: $VERSION"
echo "[*] Output: $OUT_FILE"

# ARM64 for modern Armbian (Orange Pi, Rock Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o "$OUT_FILE" .

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ BUILD SUCCESS!"
    echo "File '$OUT_FILE' siap diupload ke VPS Armbian."
else
    echo ""
    echo "❌ BUILD FAILED!"
fi
