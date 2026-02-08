@echo off
echo ========================================
echo   SmartBell Build for Linux (VPS) 🐧
echo ========================================
echo.

echo [*] Setting Environment Variables (GOOS=linux, GOARCH=amd64)...
set GOOS=linux
set GOARCH=amd64
set VERSION=v1

:: Get Date in DDMMYY format using PowerShell (Regional Safe)
for /f "tokens=*" %%a in ('powershell -Command "Get-Date -Format 'ddMMyy'"') do set DATE=%%a

set OUT_FILE=smartbell_%VERSION%_%DATE%

echo [*] Compiling binary to %OUT_FILE%...
go build -o %OUT_FILE% main.go holiday_handlers.go point_handlers.go report_handlers.go report_helpers.go report_pdf.go

if %ERRORLEVEL% equ 0 (
    echo.
    echo ✅ BUILD SUCCESS! 
    echo File '%OUT_FILE%' siap diupload ke VPS.
) else (
    echo.
    echo ❌ BUILD FAILED!
)
pause
