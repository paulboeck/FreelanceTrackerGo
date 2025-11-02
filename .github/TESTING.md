# Testing Guide

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

- **Total Tests**: 120+
- **E2E Tests**: 4 (1 skipped in CI)
- **PDF Generation Tests**: 8 (all skipped in CI)
- **Unit Tests**: ~70
- **Integration Tests**: ~50

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
