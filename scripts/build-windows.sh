#!/bin/bash
set -e

# Script to build Windows executable with manifest and resources
# Usage: ./scripts/build-windows.sh [version]

VERSION=${1:-dev}
DIST_DIR="dist"
BUILD_DIR="bin"
WINDOWS_DIR="${DIST_DIR}/windows"

echo "Building Windows package for version ${VERSION}..."

# Create directory structure
mkdir -p "${WINDOWS_DIR}"

# Check if Windows binary exists
if [ ! -f "${BUILD_DIR}/web-windows-amd64.exe" ]; then
    echo "Error: Windows binary not found. Run 'make build-windows' first."
    exit 1
fi

# Copy the binary
echo "Copying binary..."
cp "${BUILD_DIR}/web-windows-amd64.exe" "${WINDOWS_DIR}/FreelanceTracker.exe"

# Copy UI assets and migrations
echo "Copying resources..."
cp -R ui "${WINDOWS_DIR}/"
cp -R migrations "${WINDOWS_DIR}/"

# Create a batch file launcher
echo "Creating launcher script..."
cat > "${WINDOWS_DIR}/FreelanceTracker.bat" << 'EOF'
@echo off
REM FreelanceTracker Launcher for Windows

REM Get the directory where this script is located
set "APP_DIR=%~dp0"

REM Set database path to user's AppData
set "DB_DIR=%APPDATA%\FreelanceTracker"
if not exist "%DB_DIR%" mkdir "%DB_DIR%"
set "DB_PATH=%DB_DIR%\freelance_tracker.db"

REM Change to application directory
cd /d "%APP_DIR%"

REM Launch the application
echo Starting FreelanceTracker...
echo Database location: %DB_PATH%
echo Web interface will be available at: http://localhost:8080
echo.
echo Press Ctrl+C to stop the application
echo.

FreelanceTracker.exe -dsn="%DB_PATH%"
EOF

# Create a PowerShell launcher as well (more modern)
cat > "${WINDOWS_DIR}/FreelanceTracker.ps1" << 'EOF'
# FreelanceTracker Launcher for Windows (PowerShell)

$AppDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$DbDir = Join-Path $env:APPDATA "FreelanceTracker"
$DbPath = Join-Path $DbDir "freelance_tracker.db"

# Create database directory if it doesn't exist
if (-not (Test-Path $DbDir)) {
    New-Item -ItemType Directory -Path $DbDir | Out-Null
}

# Change to application directory
Set-Location $AppDir

# Launch the application
Write-Host "Starting FreelanceTracker..." -ForegroundColor Green
Write-Host "Database location: $DbPath" -ForegroundColor Cyan
Write-Host "Web interface will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C to stop the application" -ForegroundColor Yellow
Write-Host ""

& .\FreelanceTracker.exe -dsn="$DbPath"
EOF

# Create README
cat > "${WINDOWS_DIR}/README.txt" << EOF
FreelanceTracker ${VERSION} - Windows Edition

This is a Windows application for tracking freelance clients, projects,
timesheets, and invoices.

INSTALLATION:
1. Extract all files to a folder of your choice (e.g., C:\FreelanceTracker)
2. Run FreelanceTracker.bat or FreelanceTracker.exe

USAGE:
- Double-click FreelanceTracker.bat to start the application
- The web interface will be available at http://localhost:8080
- Open your web browser to http://localhost:8080
- The database will be stored in: %APPDATA%\FreelanceTracker\

STOPPING THE APPLICATION:
- Press Ctrl+C in the command window
- Or close the command window

REQUIREMENTS:
- Windows 10 or later (64-bit)
- No additional dependencies required

Version: ${VERSION}
Build Date: $(date -u '+%Y-%m-%d %H:%M:%S UTC')

For support and updates, visit: https://github.com/paulboeck/FreelanceTrackerGo
EOF

# Create a simple installer info file
cat > "${WINDOWS_DIR}/INSTALL.txt" << 'EOF'
INSTALLATION INSTRUCTIONS
=========================

QUICK START:
1. Extract all files from the ZIP archive to a folder
2. Double-click "FreelanceTracker.bat" to start
3. Open http://localhost:8080 in your web browser

DETAILED INSTALLATION:
1. Create a folder (e.g., C:\Program Files\FreelanceTracker)
2. Extract all files to that folder
3. (Optional) Create a desktop shortcut to FreelanceTracker.bat
4. Run FreelanceTracker.bat

The application will automatically create its database in:
  %APPDATA%\FreelanceTracker\freelance_tracker.db

TROUBLESHOOTING:
- If Windows Defender blocks the application, click "More info" then "Run anyway"
- Make sure no other application is using port 8080
- Check Windows Firewall settings if you can't access the web interface

UNINSTALLATION:
1. Delete the application folder
2. Delete the database folder: %APPDATA%\FreelanceTracker\
EOF

echo "✓ Windows package created: ${WINDOWS_DIR}"
echo ""
echo "To create a ZIP distribution:"
echo "  cd ${DIST_DIR} && zip -r FreelanceTracker-${VERSION}-windows.zip windows/"
echo ""
echo "Package contents:"
ls -lh "${WINDOWS_DIR}/"
