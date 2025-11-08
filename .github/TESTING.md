# Testing Guide

## Quick Reference

- [Running Tests](#running-tests) - Commands for running all test types
- [Security Tests](#security-tests) - OWASP Top 10 vulnerability testing
- [End-to-End Tests](#end-to-end-e2e-tests) - Browser automation tests
- [Test Statistics](#test-statistics) - Current test counts and coverage
- [Skipping Flaky Tests in CI](#skipping-flaky-tests-in-ci) - How to handle CI-specific issues
- [Debugging Test Failures](#debugging-test-failures) - Troubleshooting guide

## Test Files

### Handler Tests (`cmd/web/`)
- `handlers_test.go` - General HTTP handler tests
- `invoice_handlers_test.go` - Invoice-specific handler tests
- `invoice_recalculation_test.go` - Invoice recalculation logic tests
- `settings_test.go` - Settings management tests
- `timesheet_handlers_test.go` - Timesheet handler tests
- `user_edit_test.go` - User editing functionality tests

### Security Tests (`cmd/web/`)
- `security_test.go` - OWASP Top 10 security vulnerability tests

### E2E Tests (`cmd/web/`)
- `e2e_setup_test.go` - E2E test infrastructure
- `e2e_simple_test.go` - Basic navigation tests
- `e2e_login_test.go` - Authentication workflow tests

### Model Tests (`internal/models/`)
- `clients_test.go` - Client model CRUD operations
- `currency_test.go` - Currency handling tests
- `invoices_test.go` - Invoice model and PDF generation tests
- `projects_test.go` - Project model tests
- `settings_test.go` - Application settings tests
- `timesheets_test.go` - Timesheet model tests
- `users_test.go` - User model and authentication tests

### Other Tests
- `internal/crypto/aes_test.go` - AES encryption/decryption tests
- `internal/email/smtp_test.go` - Email sending tests

## Running Tests

### Locally

```bash
# Run all tests (including potentially flaky tests)
make test

# Run with verbose output
make test-verbose

# Run with race detection and coverage (matches CI minus skipped tests)
make test-ci

# Run specific test types
make test-e2e      # End-to-end browser tests
make test-unit     # Unit tests only
make test-http     # HTTP integration tests

# Run security tests only
go test -v -run TestOWASP ./cmd/web

# Run specific security test category
go test -v -run TestOWASP_SQLInjection ./cmd/web
go test -v -run TestOWASP_XSSAttacks ./cmd/web
go test -v -run TestOWASP_AuthenticationFailures ./cmd/web
```

### In CI

CI runs with the `-short` flag to skip tests that are flaky in CI environments:

```bash
go test -v -race -short -timeout=20m ./...
```

## Skipping Flaky Tests in CI

Some tests may be reliable locally but flaky in CI due to:
- Slower CI environment
- Different timing/network conditions
- Browser initialization delays

### How to Skip a Test in CI

To skip a test in CI while keeping it for local development, add this check at the start of the test:

```go
func TestSomething(t *testing.T) {
    t.Run("potentially flaky subtest", func(t *testing.T) {
        // Skip this test in CI
        if testing.Short() {
            t.Skip("Skipping in CI due to timing issues")
        }

        // Your test code here...
    })
}
```

### Currently Skipped Tests

The following tests are skipped in CI (when `-short` flag is used):

- **`TestE2EUserLogin/logout_workflow`** (`cmd/web/e2e_login_test.go`)
  - Reason: Timing issues with redirect chains in CI
  - Status: Works reliably locally
  - Issue: Browser takes ~57 seconds to start in CI, leaving little time for redirect handling

- **`TestInvoiceModel_GenerateComprehensivePDF`** (`internal/models/invoices_test.go`)
  - Reason: Chrome startup timing issues in CI (websocket url timeout)
  - Status: Works reliably locally (all 8 subtests pass)
  - Issue: chromedp WebSocket connection times out waiting for Chrome to start in CI environment

### Best Practices

1. **Always run locally without `-short` first** to ensure tests pass
2. **Document why tests are skipped** in the code comments
3. **Use skip sparingly** - only for genuinely flaky tests
4. **Consider fixing instead of skipping** when possible
5. **Mark skipped tests with a TODO** if you plan to fix them later

### Example: Marking a Test for CI Skip

```go
t.Run("flaky operation", func(t *testing.T) {
    // TODO: Fix timing issues in CI - works locally
    if testing.Short() {
        t.Skip("Skipping in CI: timing issues with external service")
    }

    // Test code...
})
```

## Test Statistics

- **Total Test Cases**: 356+
- **Security Tests**: 24 (OWASP Top 10 coverage)
- **E2E Tests**: 8 (1 skipped in CI)
- **PDF Generation Tests**: 8 (all skipped in CI)
- **HTTP Handler Tests**: ~50
- **Model Unit Tests**: ~80
- **Integration Tests**: ~186

## Security Tests

The application includes comprehensive security tests that verify protection against OWASP Top 10 vulnerabilities.

### Test Coverage

All security tests are located in `cmd/web/security_test.go` and cover:

| Test Category | Test Count | OWASP Category | Status |
|---|---|---|---|
| SQL Injection | 11 payloads | A03:2021 - Injection | ✅ Pass |
| XSS Attacks | 8 payloads | A03:2021 - Injection | ✅ Pass |
| Broken Access Control | 2 scenarios | A01:2021 | ✅ Pass |
| Authentication Failures | 3 tests | A07:2021 | ✅ Pass |
| Sensitive Data Exposure | 3 tests | A02:2021 - Cryptographic Failures | ✅ Pass |
| Input Validation | 11 edge cases | A03:2021 - Injection | ✅ Pass |
| Security Misconfiguration | 2 tests | A05:2021 | ✅ Pass |
| **Total** | **24 tests** | **7 OWASP categories** | ✅ **100% Pass** |

### Running Security Tests

```bash
# Run all security tests
go test -v -run TestOWASP ./cmd/web

# Run specific OWASP category
go test -v -run TestOWASP_SQLInjection ./cmd/web
go test -v -run TestOWASP_XSSAttacks ./cmd/web
go test -v -run TestOWASP_BrokenAccessControl ./cmd/web
go test -v -run TestOWASP_AuthenticationFailures ./cmd/web
go test -v -run TestOWASP_SensitiveDataExposure ./cmd/web
go test -v -run TestOWASP_InputValidation ./cmd/web
go test -v -run TestOWASP_SecurityMisconfiguration ./cmd/web
```

### Security Test Details

**SQL Injection Protection** (`TestOWASP_SQLInjection`):
- Tests 11 common SQL injection payloads
- Verifies parameterized queries prevent injection
- Confirms database remains stable after injection attempts

**XSS Attack Protection** (`TestOWASP_XSSAttacks`):
- Tests 8 XSS payload variants
- Verifies HTML template escaping works correctly
- Confirms malicious scripts cannot execute

**Authentication Security** (`TestOWASP_AuthenticationFailures`):
- Verifies bcrypt password hashing (cost factor 12)
- Tests that invalid credentials are rejected
- Confirms session cookies have security attributes (HttpOnly, SameSite)

**Sensitive Data Protection** (`TestOWASP_SensitiveDataExposure`):
- Confirms passwords never appear in responses
- Verifies SMTP passwords are encrypted in database (AES-256-GCM)
- Tests that error messages don't leak implementation details

**Input Validation** (`TestOWASP_InputValidation`):
- Tests extremely long inputs (10,000+ characters)
- Validates null byte handling
- Tests path traversal attempts
- Validates command injection prevention
- Tests template injection prevention

For complete security documentation, see [SECURITY.md](./SECURITY.md).

### Security Test Automation

Security tests run automatically:
- ✅ With `make test`
- ✅ With `go test ./...`
- ✅ In CI/CD pipeline
- ✅ In coverage reports

No special configuration needed - they're part of the standard test suite.

## End-to-End (E2E) Tests

E2E tests use Rod for browser automation to test complete user workflows. These tests require Chrome/Chromium.

### Test Files

- `cmd/web/e2e_setup_test.go` - E2E test infrastructure and setup
- `cmd/web/e2e_simple_test.go` - Basic navigation and page load tests
- `cmd/web/e2e_login_test.go` - User authentication workflows

### Running E2E Tests

```bash
# Run all E2E tests
make test-e2e

# Or directly with go test
go test -v -run TestE2E ./cmd/web -timeout=60s

# Run specific E2E test
go test -v -run TestE2ESimple ./cmd/web
go test -v -run TestE2EUserLogin ./cmd/web
```

### E2E Test Coverage

| Test | Description | Status |
|---|---|---|
| `TestE2ESimple/home_page_loads` | Verifies home page loads correctly | ✅ Pass |
| `TestE2ESimple/clients_page_loads` | Tests client list page | ✅ Pass |
| `TestE2ESimple/projects_page_loads` | Tests projects page | ✅ Pass |
| `TestE2EUserLogin/successful_login_workflow` | Full login flow | ✅ Pass |
| `TestE2EUserLogin/failed_login_shows_error` | Invalid credentials | ✅ Pass |
| `TestE2EUserLogin/logout_workflow` | Logout functionality | ⚠️ Skipped in CI |

### Requirements

**Local Development:**
- Chrome or Chromium browser (auto-detected)
- Browser will launch in headless mode during tests

**CI Environment:**
- Chrome installed via `browser-actions/setup-chrome@v1`
- Runs in headless mode with `--no-sandbox` flag

### E2E Test Timeouts

E2E tests have longer timeouts due to browser startup:
- Default test timeout: 60 seconds
- Browser initialization: ~5-10 seconds locally, ~30-57 seconds in CI
- Page navigation: 5-10 seconds per operation

### Known Issues

- **Logout workflow test**: Skipped in CI due to redirect chain timing issues
  - Works reliably locally
  - CI environment is ~10x slower causing timeout issues

## Debugging Test Failures

### Local Failures

```bash
# Run specific test with verbose output
go test -v -run TestName ./path/to/package

# Run with race detector
go test -race -run TestName ./path/to/package

# Run multiple times to catch flaky tests
go test -count=10 -run TestName ./path/to/package
```

### CI Failures

1. Check if test is already marked to skip with `-short`
2. Check Chrome installation in CI logs (we verify Chrome is installed)
3. Check timing - CI is ~10x slower than local
4. Consider adding the `-short` skip if consistently flaky

### E2E Test Specific Issues

E2E tests use Rod for browser automation and require Chrome:

- **Local**: Chrome should be auto-detected
- **CI**: Chrome is installed via `browser-actions/setup-chrome@v1`
- **Timeouts**: E2E tests have 60s timeout by default
- **Headless**: Tests run in headless mode with `--no-sandbox` for CI

## Coverage

Generate coverage report locally:

```bash
make test-coverage
open coverage.html
```

CI automatically uploads coverage to Codecov on every PR.
