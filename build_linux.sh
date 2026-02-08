#!/bin/bash

echo "========================================"
echo "   SmartBell Build for Linux (VPS) 🐧"
echo "========================================"
echo ""

echo "[*] Compiling binary for Linux..."

VERSION="v1"
DATE=$(date +"%d%m%y")
OUT_FILE="smartbell_${VERSION}_${DATE}"

GOOS=linux GOARCH=amd64 go build -o $OUT_FILE main.go holiday_handlers.go point_handlers.go report_handlers.go report_helpers.go report_pdf.go

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ BUILD SUCCESS!"
    echo "File '$OUT_FILE' siap diupload ke VPS."
else
    echo ""
    echo "❌ BUILD FAILED!"
fi
