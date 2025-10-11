#!/bin/bash
set -e

# Script to build macOS .app bundle for FreelanceTracker
# Usage: ./scripts/build-macos-app.sh [version]

VERSION=${1:-dev}
APP_NAME="FreelanceTracker"
DIST_DIR="dist"
BUILD_DIR="bin"
APP_DIR="${DIST_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${APP_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "Building macOS application bundle for version ${VERSION}..."

# Clean and create directory structure
rm -rf "${APP_DIR}"
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# Copy the universal binary
echo "Copying binary..."
cp "${BUILD_DIR}/web-darwin-universal" "${MACOS_DIR}/${APP_NAME}"
chmod +x "${MACOS_DIR}/${APP_NAME}"

# Copy UI assets and migrations
echo "Copying resources..."
cp -R ui "${RESOURCES_DIR}/"
cp -R migrations "${RESOURCES_DIR}/"

# Create Info.plist
echo "Creating Info.plist..."
cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>com.freelancetracker.app</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSUIElement</key>
    <false/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2025 FreelanceTracker. All rights reserved.</string>
</dict>
</plist>
EOF

# Create a launcher script that sets up the environment
echo "Creating launcher script..."
cat > "${MACOS_DIR}/${APP_NAME}-launcher" << 'EOF'
#!/bin/bash
# Launcher script for FreelanceTracker

# Get the directory containing this script (the MacOS directory)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
RESOURCES_DIR="${DIR}/../Resources"

# Change to resources directory so relative paths work
cd "${RESOURCES_DIR}"

# Set database path to user's Application Support
DB_DIR="${HOME}/Library/Application Support/FreelanceTracker"
mkdir -p "${DB_DIR}"
DB_PATH="${DB_DIR}/freelance_tracker.db"

# Launch the application
exec "${DIR}/FreelanceTracker" -dsn="${DB_PATH}"
EOF

chmod +x "${MACOS_DIR}/${APP_NAME}-launcher"

# Create PkgInfo
echo "APPL????" > "${CONTENTS_DIR}/PkgInfo"

# Create a simple README for the app
cat > "${RESOURCES_DIR}/README.txt" << EOF
FreelanceTracker ${VERSION}

This is a macOS application for tracking freelance clients, projects, timesheets, and invoices.

The application will store its database in:
~/Library/Application Support/FreelanceTracker/

To run the application:
1. Double-click FreelanceTracker.app
2. The web interface will be available at http://localhost:8080
3. Open your web browser to http://localhost:8080

Version: ${VERSION}
EOF

echo "✓ macOS application bundle created: ${APP_DIR}"
echo ""
echo "To test the application:"
echo "  open ${APP_DIR}"
echo ""
echo "To create a DMG for distribution:"
echo "  hdiutil create -volname \"FreelanceTracker\" -srcfolder ${APP_DIR} -ov -format UDZO ${DIST_DIR}/FreelanceTracker-${VERSION}.dmg"
