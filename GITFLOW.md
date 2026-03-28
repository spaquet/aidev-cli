# AIDev CLI - GitFlow CI/CD Process

This document describes the automated CI/CD pipeline for building and releasing the AIDev CLI across multiple platforms and architectures.

## Overview

The gitflow consists of two main workflows:

1. **CI Workflow** (`.github/workflows/ci.yml`) — Runs on every push to `main` and pull requests
2. **Release Workflow** (`.github/workflows/release.yml`) — Runs when a version tag is pushed

## CI Workflow

### Trigger
- Push to `main` branch
- Pull requests to `main` branch

### Jobs

#### 1. **Lint** (Hard Gate)
- Runs on: `ubuntu-latest`
- Tool: `golangci-lint v1.59.0`
- **Blocks merges if it fails**
- Duration: ~1 min

#### 2. **Test Generic** (Multi-platform)
- Runs on: `ubuntu-latest`, `macos-latest`, `windows-latest`
- Steps:
  - `go vet ./...` — Code quality checks
  - Build binary — Smoke test to ensure code compiles
  - `go test -v -cover ./...` — All unit tests
- Duration: ~2 min per platform

#### 3. **Test Platform-Specific**
Three separate jobs for OS-specific concerns:

| Job | Platform | Purpose |
|-----|----------|---------|
| `test-platform-linux` | Ubuntu | Tests `syscall.WaitStatus` (SSH exit codes), Unix signals |
| `test-platform-windows` | Windows | Tests `archive extraction`, `unzip` availability, path handling |
| `test-platform-darwin` | macOS | Tests XDG config paths, macOS-specific compilation |

## Release Workflow

### Trigger
When a version tag is pushed:
```bash
git tag v0.2.0
git push origin v0.2.0
```

### Pre-Release Validation
1. All CI tests re-run on all platforms (must pass)
2. Linting check runs (must pass)
3. If any test fails, the release is **blocked** and does not proceed

### Release Job
1. Runs `goreleaser release --clean`
2. Builds binaries for all target platforms/architectures:
   - Linux: amd64, arm64
   - macOS: amd64, arm64
   - Windows: amd64
3. Creates archives (.tar.gz for Unix, .zip for Windows)
4. Generates checksums (SHA256)
5. Publishes to GitHub Releases with changelog

### Artifacts Published
```
aidev_v0.2.0_linux_amd64.tar.gz
aidev_v0.2.0_linux_arm64.tar.gz
aidev_v0.2.0_darwin_amd64.tar.gz
aidev_v0.2.0_darwin_arm64.tar.gz
aidev_v0.2.0_windows_amd64.zip
checksums.txt
```

## Supported Platforms & Architectures

| OS | Architectures | Notes |
|---|---|---|
| Linux | amd64, arm64 | Full support for CLI and TUI |
| macOS | amd64, arm64 | Full support; universal binary via GoReleaser |
| Windows | amd64 | Limited: no native SSH, requires WSL for full features |

## Test Structure

### Generic Tests (All Platforms)
Located in:
- `internal/models/instance_test.go` — Model struct tests
- `internal/commands/update_test.go` — Archive format detection
- `internal/auth/store_test.go` — Config file read/write

### Platform-Specific Tests
- `internal/ssh/ssh_test.go` — Unix-only (uses build tag `//go:build linux || darwin`)

### Test Helpers
- `internal/testutil/testutil.go` — Shared utilities
  - `TempConfigDir(t)` — Creates isolated temp config directory
  - `SkipIfNoSSH(t)` — Skips tests if SSH binary not available

## Local Testing

### Run all tests locally
```bash
make test
```

### Run tests for a specific platform
```bash
# Linux tests only
go test -v -tags linux -run '.*Linux.*' ./...

# Windows tests only
go test -v -tags windows -run '.*Windows.*' ./...

# macOS tests only
go test -v -tags darwin -run '.*Darwin.*' ./...
```

### Build for all platforms locally
```bash
make cross-build
```

### Lint locally
```bash
# Requires golangci-lint: brew install golangci-lint
golangci-lint run ./...
```

## Creating a Release

### 1. Prepare the Release
```bash
# Update version in code if needed (go.mod version is auto-detected from git tags)
# Run all tests locally
make test
make lint  # if available
```

### 2. Create the Tag
```bash
# Ensure you're on main and fully up to date
git checkout main
git pull origin main

# Create an annotated tag (preferred)
git tag -a v0.2.0 -m "Release v0.2.0: Feature X, Bug fix Y"

# Or lightweight tag
git tag v0.2.0
```

### 3. Push the Tag
```bash
git push origin v0.2.0
```

### 4. Monitor the Release
- GitHub Actions will automatically run the release workflow
- Check the "Actions" tab in GitHub to watch progress
- Once complete, artifacts will appear in GitHub Releases

### 5. Verify Release
- Visit GitHub Releases page
- Verify all artifacts are present
- Download and test binaries on different platforms if possible

## Troubleshooting

### Tests fail on a specific platform
1. Check the workflow logs in GitHub Actions
2. Reproduce locally by testing on that platform
3. Fix the issue, push, and the CI will re-run

### Release doesn't trigger
- Ensure the tag format matches `v*.*.*` (semantic versioning)
- Verify the tag is pushed to origin: `git push origin v0.2.0`
- Check the tag exists: `git tag -l v0.2.0`

### GoReleaser fails
- GoReleaser configuration is in `.goreleaser.yml`
- Common issues:
  - Missing `.git` metadata: ensure `.git` directory is present (GitHub Actions handles this)
  - Archive format mismatches: check OS-specific settings in config

## Continuous Improvement

### Adding more tests
1. Create `*_test.go` files in the package
2. Run `go test -v ./...` to verify
3. Commit and push — CI will run automatically

### Platform-specific fixes
1. Use build tags: `//go:build linux || darwin`
2. Create conditional code paths based on `runtime.GOOS`
3. Test locally with `GOOS=<target> GOARCH=<target> go test ./...`

## References

- [GoReleaser Docs](https://goreleaser.com/)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Go Testing](https://golang.org/doc/effective_go#testing)
