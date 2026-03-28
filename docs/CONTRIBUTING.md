# Contributing to AIDev CLI

This guide covers development workflow, testing, and the release process for the AIDev CLI project.

## Table of Contents

- [Development Setup](#development-setup)
- [Testing](#testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Creating a Release](#creating-a-release)
- [Workflow](#workflow)

## Development Setup

### Prerequisites

- Go 1.23+
- Git
- `ssh` binary (for SSH tests)
- `golangci-lint` (optional, for linting)

### Build Locally

```bash
# Clone the repository
git clone https://github.com/aidev/aidev-cli.git
cd aidev-cli

# Build the binary
go build -o aidev ./cmd/aidev

# Run it
./aidev --version
```

### Development Build with Makefile

```bash
# Build to bin/
make build

# Run the TUI
make run

# Build and run login command
make run-login

# Clean build artifacts
make clean
```

## Testing

### Run All Tests

```bash
make test
```

This runs `go test -v -cover ./...` on all packages.

### Test a Specific Package

```bash
go test -v ./internal/auth
go test -v ./internal/commands
```

### Test with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View in browser
```

### Test Structure

**Unit Tests:**
- Located in `*_test.go` files alongside source code
- Test models, config storage, platform-specific logic
- No external dependencies (use temp dirs for file I/O)

**Platform-Specific Tests:**
- Some tests use build tags: `//go:build linux || darwin`
- Windows-specific code tested on windows runner
- SSH tests skip if `ssh` binary not available

### Example: Testing the Auth Store

```bash
go test -v ./internal/auth -run TestStore_SaveAndLoad
```

## CI/CD Pipeline

### GitHub Actions Workflows

Two workflows automate testing and releases:

#### 1. CI Workflow (`.github/workflows/ci.yml`)

Triggers on: `push` to `main`, `pull_request` to `main`

**Jobs:**

- **Lint** (Ubuntu only, hard gate)
  - golangci-lint v1.59.0
  - Must pass to merge PRs
  - ~1 min

- **Test Generic** (Matrix: Ubuntu, macOS, Windows)
  - `go vet ./...`
  - `go build` (smoke test)
  - `go test -v -cover ./...`
  - ~2 min per platform

- **Test Platform-Specific** (3 separate jobs)
  - Linux: Tests for `syscall.WaitStatus`
  - Windows: Tests for archive extraction, paths
  - macOS: Tests for XDG config, compilation

**Example PR Workflow:**
```
1. Push commit to branch
2. GitHub Actions runs:
   - lint (fails? PR blocked)
   - test-generic (all 3 platforms, all must pass)
   - test-platform-* (OS-specific tests)
3. All tests pass → PR ready to merge
```

#### 2. Release Workflow (`.github/workflows/release.yml`)

Triggers on: `push` to tags matching `v*.*.*`

**Process:**

1. **Pre-Release Tests** (all platforms)
   - Re-run all CI tests (must all pass)
   - If any test fails → release blocked

2. **GoReleaser** (Ubuntu only)
   - Builds all platform/arch combinations
   - Creates archives (.tar.gz, .zip)
   - Generates SHA256 checksums
   - Creates GitHub Release with download links
   - (Optional) Publishes to Homebrew tap

**Release Artifacts:**
```
aidev_v0.2.0_linux_amd64.tar.gz
aidev_v0.2.0_linux_arm64.tar.gz
aidev_v0.2.0_darwin_amd64.tar.gz
aidev_v0.2.0_darwin_arm64.tar.gz
aidev_v0.2.0_windows_amd64.zip
checksums.txt
```

### Local Testing (Before Pushing)

Before committing:

```bash
# Lint (requires golangci-lint installed)
make lint
# or
golangci-lint run ./...

# Test all packages
make test

# Build for all platforms
make cross-build
```

If anything fails, fix it locally before pushing.

## Creating a Release

### Step 1: Prepare the Release

Ensure your code is ready:

```bash
# Update version/changelog in code if needed (optional)
# Version is auto-detected from git tags

# Run all tests locally
make test

# Lint check
make lint

# Verify cross-platform builds
make cross-build
```

### Step 2: Create the Git Tag

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create an annotated tag (preferred)
git tag -a v0.2.0 -m "Release v0.2.0: Feature X, Bug fix Y"

# Or lightweight tag
git tag v0.2.0
```

**Tag Format:** Must match `v*.*.*` (semantic versioning)
- ✅ v0.2.0
- ✅ v1.0.0-rc1
- ❌ 0.2.0 (missing v prefix)
- ❌ release-0.2.0 (wrong format)

### Step 3: Push the Tag

```bash
# Push the tag to GitHub
git push origin v0.2.0

# Or push all tags
git push origin --tags
```

This triggers the release workflow.

### Step 4: Monitor the Release

```bash
# Watch the workflow in GitHub UI
# Go to: https://github.com/aidev/aidev-cli/actions
# Click the "Release" workflow run
```

Or via GitHub CLI:

```bash
gh run list --workflow=release.yml
gh run view <run-id> --log
```

### Step 5: Verify the Release

Once the workflow completes:

1. Check GitHub Releases page: https://github.com/aidev/aidev-cli/releases
2. Verify all artifacts are present (5 binaries + checksums.txt)
3. Download and test a binary on the target platform if possible

**Example verification:**

```bash
# Download macOS binary
curl -LO https://github.com/aidev/aidev-cli/releases/download/v0.2.0/aidev_v0.2.0_darwin_amd64.tar.gz

# Extract and test
tar xzf aidev_v0.2.0_darwin_amd64.tar.gz
./aidev/aidev --version
```

## Workflow

### Feature Development

1. Create a branch from `main`:
   ```bash
   git checkout -b feature/new-feature
   ```

2. Make changes and commit (use conventional commits):
   ```bash
   git commit -m "feat: add new feature"
   git commit -m "fix: resolve issue"
   ```

3. Push and create a PR:
   ```bash
   git push origin feature/new-feature
   ```

4. GitHub Actions will automatically:
   - Run linting (hard gate)
   - Run tests on 3 platforms
   - Run platform-specific tests

5. Merge PR once all checks pass

### Hotfix

If you need to fix a critical bug in a released version:

1. Create a hotfix branch:
   ```bash
   git checkout -b hotfix/critical-fix main
   ```

2. Make the fix and commit:
   ```bash
   git commit -m "fix: critical issue"
   ```

3. Create a PR, merge once tests pass

4. Tag and release:
   ```bash
   git tag -a v0.2.1 -m "Hotfix: description"
   git push origin v0.2.1
   ```

### Documentation Updates

Documentation is expected to stay in sync with code:

- When implementing features: update `docs/ARCHITECTURE.md`
- When changing API integration: update `docs/rails-api-spec.md`
- When changing TUI: update `docs/tui-design.md`
- When changing config: update `README.md`

## Code Style

### Formatting

```bash
# Format all code
make fmt
# or
go fmt ./...
```

### Linting

```bash
make lint
# or
golangci-lint run ./...
```

### Conventions

- Package names: lowercase, single word (`api`, `auth`, `tui`)
- Type names: PascalCase (`LoginModel`, `Config`, `Instance`)
- Function names: camelCase (`GetInstances`, `SaveConfig`)
- Constants: UPPERCASE (`MaxRetries`, `DefaultTimeout`)
- Comments: Full sentences, capital letters

### Testing

- Test file format: `filename_test.go`
- Test function format: `func TestFeatureName(t *testing.T)`
- Use table-driven tests for multiple cases
- Use testutil helpers for setup (temp dirs, mocks)

## Troubleshooting

### Tests fail locally but pass on GitHub

Usually due to:
- Different Go version (use `1.23+`)
- Missing dependencies (`go mod tidy`)
- Platform-specific issues (test on actual platform)

Solution:
```bash
go version  # Check version
go mod tidy
go test ./...
```

### Lint fails but I don't know why

```bash
# Lint with verbose output
golangci-lint run -v ./...

# Check specific linter
golangci-lint run --linters=golint ./...
```

### Release doesn't trigger

- Tag format must match `v*.*.*`
- Confirm tag was pushed: `git tag -l v0.2.0`
- Verify push: `git push --tags` or `git push origin v0.2.0`
- Check GitHub Actions tab for errors

### Binary size is too large

Binary is expected to be ~10 MB. If it's larger:

```bash
# Check what's in the binary
go build -ldflags="-s -w" ./cmd/aidev  # Strip debug symbols
ls -lh aidev
```

The `.goreleaser.yml` uses `-s -w` flags automatically.

## Release Checklist

Before creating a release tag:

- [ ] All tests pass locally (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Cross-platform builds succeed (`make cross-build`)
- [ ] Changelog/release notes are ready
- [ ] Version is documented if needed
- [ ] No uncommitted changes (`git status` is clean)

## Questions?

- Check the [Architecture](ARCHITECTURE.md) guide for system design
- See [README.md](../README.md) for user documentation
- Open an issue on GitHub for questions about the codebase

## Related Documentation

- [Architecture](ARCHITECTURE.md) — System design and components
- [CI/CD Pipeline Details](../GITFLOW.md) — Detailed pipeline explanation
- [TUI Design](tui-design.md) — UI/UX specification
- [API Reference](rails-api-spec.md) — Backend API details
