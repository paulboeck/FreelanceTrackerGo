# FreelanceTrackerGo

A Go web application for tracking freelance clients, projects, timesheets, and invoices.

## Quick Start

```bash
# Run the application
go run ./cmd/web

# Access the application
open http://localhost:8080
```

The application will automatically create a database at `~/FreelanceTracker/freelance_tracker.db`.

**Database Location:**
- By default, the database is stored in `~/FreelanceTracker/` on all platforms
- You can override this with the `-dsn` flag: `go run ./cmd/web -dsn="/path/to/database.db"`
- The application will automatically create the directory if it doesn't exist

## Email Configuration

FreelanceTrackerGo can send invoices directly to clients via email using Gmail SMTP. Follow these steps to set up email functionality:

### Step 1: Enable Gmail App Passwords

1. **Enable 2-Step Verification on your Gmail account:**
   - Go to [Google Account Settings](https://myaccount.google.com/)
   - Click "Security" in the left sidebar
   - Under "How you sign in to Google", find "2-Step Verification"
   - Click "Get started" and follow the on-screen instructions
   - Click "Turn on" to enable 2-Step Verification

2. **Generate an App Password:**
   - Go directly to [App Passwords](https://myaccount.google.com/apppasswords)
   - You may need to sign in again to verify your identity
   - Enter an app name (e.g., "FreelanceTracker") in the text field
   - Click "Create"
   - **Important**: Copy the 16-character password that appears (no spaces)
   - **Note**: You won't see this password again, so store it securely

**Important Notes:**
- App passwords are only available if 2-Step Verification is enabled
- For work/school accounts, app passwords may be disabled by your administrator
- The 16-character password replaces your regular Gmail password for this application

### Step 2: Configure FreelanceTrackerGo

1. **Start the application:**
   ```bash
   go run ./cmd/web
   ```

2. **Navigate to Settings:**
   - Open http://localhost:4000 in your browser
   - Click "Settings" in the navigation menu
   - Click "Edit Settings"

3. **Configure Email Settings:**
   - Set **Email Enabled** to "Yes"
   - Enter your **SMTP Username**: Your full Gmail address (e.g., `your.email@gmail.com`)
   - Enter your **SMTP Password**: The 16-character app password from Step 1
   - Leave other settings as defaults:
     - **SMTP Host**: `smtp.gmail.com`
     - **SMTP Port**: `587`
     - **From Name**: `FreelanceTracker`
     - **Use TLS**: `Yes`

4. **Save Settings:**
   - Click "Update Settings"
   - You should see a success message

### Step 3: Send Invoice Emails

1. **Ensure clients have email addresses:**
   - Go to "Clients" and edit each client
   - Make sure the "Email" field is filled in

2. **Send an invoice email:**
   - Navigate to a project page
   - In the "Invoices" section, click the ✉ button next to any invoice
   - The email will be sent automatically to the client's email address
   - Success/error messages will appear at the top of the page

### Troubleshooting Email Issues

**"Authentication failed" error:**
- Double-check that 2-Factor Authentication is enabled on your Gmail account
- Verify you're using the app password, not your regular Gmail password
- Make sure the app password was copied correctly (no extra spaces)

**"Client email not found" error:**
- Edit the client and add their email address

**"Failed to connect to SMTP server" error:**
- Check your internet connection
- Verify SMTP settings match the defaults above
- Try generating a new app password

**Email not received by client:**
- Check the client's spam/junk folder
- Verify the client's email address is correct

## Authentication

Authentication is currently disabled. To re-enable:
1. Remove `display:none` property of nav div in main.css
2. Uncomment `IsAuthenticated` in helpers.go.

## Development

FreelanceTrackerGo uses a modern, reproducible build system with Make, Docker, and GitHub Actions.

### Quick Reference

```bash
# Show all available commands
make help

# Run tests
make test

# Build the application
make build

# Run the application
make run

# Generate code from SQL
make generate

# Run database migrations
make migrate
```

### Available Make Commands

**Basic Commands:**
- `make build` - Build the application binary
- `make build-release` - Build optimized release binary with version info
- `make test` - Run all tests
- `make test-verbose` - Run tests with verbose output
- `make test-coverage` - Run tests with coverage report
- `make clean` - Clean build artifacts and test databases
- `make run` - Run the application (localhost:8080)
- `make run-dev` - Run with development database
- `make migrate` - Run database migrations
- `make migrate-status` - Show migration status
- `make generate` - Generate code from SQL queries using sqlc
- `make install-tools` - Install development tools (goose, sqlc)
- `make tidy` - Tidy and verify go.mod dependencies
- `make all` - Run full build pipeline (clean, tidy, generate, test, build)

**Cross-Platform Builds:**
- `make build-macos` - Build macOS universal binary (Intel + Apple Silicon)
- `make build-macos-intel` - Build macOS Intel-only (optimized, smaller)
- `make build-macos-arm` - Build macOS Apple Silicon-only (optimized, smaller)
- `make build-macos-app` - Build macOS .app bundle
- `make build-windows` - Build Windows x86_64 (default)
- `make build-windows-amd64` - Build Windows x86_64 (explicit)
- `make build-windows-arm64` - Build Windows ARM64 (Surface Pro X, etc.)
- `make build-linux` - Build Linux x86_64 (default)
- `make build-linux-amd64` - Build Linux x86_64 (explicit)
- `make build-linux-arm64` - Build Linux ARM64 (Raspberry Pi, AWS Graviton)
- `make build-linux-arm` - Build Linux ARM 32-bit (older Raspberry Pi)
- `make build-all-platforms` - Build default for all platforms + Docker
- `make build-all-specific` - Build all platform-specific optimized binaries

**Distribution Packaging:**
- `make package-macos-app` - Create macOS .dmg installer
- `make package-windows` - Create Windows .zip package
- `make package-linux` - Create Linux .tar.gz package
- `make package-all` - Create distribution packages for all platforms

### Cross-Platform Builds

FreelanceTrackerGo supports building for multiple platforms with optimized, platform-specific binaries.

**Building for macOS:**
```bash
# Build universal binary (Intel + Apple Silicon) - recommended for distribution
make build-macos

# Or build platform-specific optimized binaries:
make build-macos-intel      # Intel Macs only (smaller file size)
make build-macos-arm        # Apple Silicon only (M1/M2/M3)

# Build complete macOS .app bundle (uses universal binary)
make build-macos-app

# Create DMG installer for distribution
make package-macos-app
```

The macOS .app bundle includes:
- Universal binary supporting both Intel and Apple Silicon
- All UI assets and database migrations
- Automatic database setup in ~/Library/Application Support/FreelanceTracker/
- Double-clickable application

**Building for Windows:**
```bash
# Build for x86_64 (standard Windows PCs)
make build-windows-amd64

# Build for ARM64 (Windows on ARM devices like Surface Pro X)
make build-windows-arm64

# Or use the default (builds x86_64)
make build-windows

# Create Windows package with launcher scripts
make package-windows
```

The Windows package includes:
- 64-bit executable optimized for your target architecture
- Batch file launcher for easy startup
- PowerShell launcher (modern alternative)
- All UI assets and database migrations
- Automatic database setup in %APPDATA%\FreelanceTracker\

**Building for Linux:**
```bash
# Build for x86_64 (standard servers/desktops)
make build-linux-amd64

# Build for ARM64 (Raspberry Pi 4/5, AWS Graviton, etc.)
make build-linux-arm64

# Build for ARM (32-bit Raspberry Pi 3 and older)
make build-linux-arm

# Or use the default (builds x86_64)
make build-linux

# Create Linux package
make package-linux
```

**Requirements for cross-compilation:**
- **Windows builds**: Requires mingw-w64 cross-compiler
  ```bash
  # macOS (via Homebrew)
  brew install mingw-w64

  # For ARM64 Windows builds, also install:
  brew install mingw-w64     # Includes aarch64 compiler

  # Linux (Ubuntu/Debian)
  sudo apt-get install mingw-w64 gcc-mingw-w64-x86-64 gcc-mingw-w64-i686
  ```

- **Linux ARM builds**: Requires ARM cross-compilation toolchain
  ```bash
  # macOS (via Homebrew)
  brew tap messense/macos-cross-toolchains
  brew install aarch64-unknown-linux-gnu arm-unknown-linux-gnueabihf

  # Linux (Ubuntu/Debian)
  sudo apt-get install gcc-aarch64-linux-gnu gcc-arm-linux-gnueabihf
  ```

- **macOS builds**: Best on macOS, but can use cross-compilation tools on Linux

**Building everything at once:**
```bash
# Build default binaries for all platforms (universal macOS, x86_64 Windows/Linux, Docker)
make build-all-platforms

# Build all platform-specific optimized binaries (no universal binaries)
make build-all-specific

# Create distribution packages for all platforms
make package-all
```

**Why choose platform-specific builds?**
- **Smaller file sizes**: Platform-specific binaries are typically 30-40% smaller than universal binaries
- **Better optimization**: Compiler can optimize specifically for the target architecture
- **Faster startup**: No need to select the correct architecture at runtime
- **Use case**: Best for controlled deployments where you know the target platform

**When to use universal/default builds?**
- **macOS universal binary**: Best for general distribution when you don't know if users have Intel or Apple Silicon
- **Default builds** (`make build-macos`, `make build-windows`, etc.): Convenient shortcuts that build the most common architecture
- **Docker**: Platform-agnostic container that runs anywhere

### Docker

Build and run using Docker:

```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Or use docker-compose
make docker-compose-up
make docker-compose-logs
make docker-compose-down
```

The Docker setup:
- Uses multi-stage builds for minimal image size
- Excludes test files from production builds
- Runs as non-root user for security
- Persists database in mounted volume

### Testing

FreelanceTrackerGo has comprehensive test coverage including unit tests, HTTP integration tests, and end-to-end browser tests.

#### Running Tests

```bash
# Run all tests
make test

# Run all tests with verbose output
make test-verbose

# Run tests with race detection and coverage (matches CI environment)
make test-ci

# Generate HTML coverage report
make test-coverage
```

#### Test Types

The project includes four types of tests:

**Security Tests** - OWASP Top 10 vulnerability testing:
```bash
# Run all security tests
go test -v -run TestOWASP ./cmd/web

# Run specific security test category
go test -v -run TestOWASP_SQLInjection ./cmd/web
```

Security tests verify protection against:
- SQL injection attacks (11 payloads tested)
- Cross-site scripting (XSS) attacks (8 payloads tested)
- Broken access control
- Authentication failures
- Sensitive data exposure
- Input validation vulnerabilities
- Security misconfiguration

**Unit Tests** - Test individual components in isolation:
```bash
make test-unit
```

**HTTP Integration Tests** - Test HTTP handlers with real requests:
```bash
make test-http
```

**End-to-End (E2E) Tests** - Full browser automation tests using Rod:
```bash
make test-e2e
```

E2E tests require Chrome/Chromium to be installed on your system. The tests:
- Start a real HTTP server on port 9876
- Use an isolated SQLite test database
- Interact with the actual UI using a headless browser
- Automatically clean up all resources after completion

#### Test Statistics

The project currently has **356+ automated test cases** covering:
- **Security**: 24 tests verifying OWASP Top 10 protections
- **E2E Browser Tests**: 8 full user workflow tests
- **Client Management**: CRUD operations and validation
- **Project Management**: Complete project lifecycle workflows
- **Timesheet Tracking**: Time entry and validation
- **Invoice Generation**: PDF creation, calculations, and email delivery
- **User Authentication**: Login, password hashing, and session security
- **Settings Management**: Application configuration and encryption
- **HTTP Handlers**: ~50 integration tests for all endpoints
- **Model Unit Tests**: ~80 tests for database operations

For detailed testing documentation, see [.github/TESTING.md](.github/TESTING.md) and [.github/SECURITY.md](.github/SECURITY.md).

#### Continuous Integration

Tests run automatically on GitHub Actions for:
- Every push to the `main` branch
- All pull requests

The CI environment:
- Installs Chrome for e2e tests
- Runs all tests with race detection (`-race` flag)
- Generates code coverage reports
- Uploads coverage to Codecov
- Cleans up test artifacts and orphaned processes

See `.github/workflows/test.yml` for the full CI configuration.

#### Test Cleanup

Tests automatically clean up after themselves, but you can manually clean test artifacts:

```bash
# Clean build artifacts and test databases
make clean

# This removes:
# - Built binaries
# - Test database files (test_*.db)
# - Coverage reports
# - Orphaned test processes
```

**Note:** Tests use SQLite with isolated test databases and automatically clean up resources after completion.

### Database Migrations

```bash
# Run migrations
make migrate

# Check migration status
make migrate-status

# Manual migration (if needed)
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 ./freelance_tracker.db up
```

Migrations run automatically on application startup.

### Code Generation

```bash
# Generate type-safe Go code from SQL queries
make generate

# Or use sqlc directly
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

Run this command whenever you modify SQL queries in the `queries/` directory.

### Build Tools

Development tools (goose, sqlc) are version-pinned in `tools.go` and managed via Go modules:

```bash
# Install tools locally
make install-tools

# Tools are also available via go run
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations sqlite3 ./freelance_tracker.db up
```

### Continuous Integration

The project uses GitHub Actions for automated testing on every push and pull request:

- **Test Job**: Runs all tests with race detection and coverage
- **Build Job**: Builds optimized binary with version info
- **Lint Job**: Runs golangci-lint for code quality
- **Docker Job**: Tests Docker build process

See `.github/workflows/test.yml` for details.

### Reproducible Builds

All builds are reproducible using:

1. **Go Modules**: Exact dependency versions in `go.mod` and `go.sum`
2. **Pinned Tools**: Build tools versioned in `tools.go`
3. **Version Info**: Git-based versioning in binaries
4. **Docker**: Containerized builds with locked dependencies

Build artifacts include version and build time:
```bash
VERSION=$(git describe --tags --always --dirty)
make build-release
```