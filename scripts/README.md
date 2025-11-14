# Development Scripts

This directory contains scripts to help with development workflow, testing, and Git hook management.

## Available Scripts

### 1. Development Tools Setup

**Purpose:** Installs all required tools for the enhanced pre-commit hook.

**Unix/Linux/Mac/Git Bash:**
```bash
bash scripts/setup-dev-tools.sh
```

**Windows Command Prompt:**
```cmd
scripts\setup-dev-tools.bat
```

**What it installs:**
- `goimports` - Code formatting and import organization (latest)
- `golangci-lint` - Comprehensive linting tool (latest)
- `gosec` - Security vulnerability scanner (latest)

**Features:**
- ✅ Checks if tools are already installed
- ✅ Shows installation progress with colored output
- ✅ Verifies installations after completion
- ✅ Provides helpful error messages
- ✅ Detects if GOBIN is in PATH
- ✅ Provides summary of installed/skipped/failed tools

**Example output:**
```
🔧 Kisanlink ERP - Development Tools Setup
================================================

This script will install the following tools:
  1. goimports  - Code formatting & import organization
  2. golangci-lint - Fast linters runner
  3. gosec - Security vulnerability scanner

1️⃣  Installing goimports... ✓ (installed v0.15.0)
2️⃣  Installing golangci-lint... ✓ (installed v1.55.2)
3️⃣  Installing gosec... ✓ (installed v2.18.2)

================================================
📊 Installation Summary

  ✅ Installed: 3
  ⏭️  Skipped:   0
  ❌ Failed:    0

✅ All tools are ready!

Next steps:
  1. Install Git hooks: bash scripts/install-hooks.sh
  2. Test pre-commit: git add . && git commit -m "test"
```

---

### 2. Git Hooks Installation

**Purpose:** Installs Git hooks from `.git-hooks/` to `.git/hooks/`.

**Unix/Linux/Mac/Git Bash:**
```bash
bash scripts/install-hooks.sh
```

**Windows Command Prompt:**
```cmd
scripts\install-hooks.bat
```

**What it does:**
- Copies `.git-hooks/pre-commit` to `.git/hooks/pre-commit`
- Makes the hook executable (Unix only)
- Verifies installation

**Example output:**
```
📦 Installing Git hooks...

✅ Installed pre-commit hook

✨ Git hooks installation complete!

The following hooks are now active:
  - pre-commit: Comprehensive quality checks before every commit

To skip hooks when committing, use: git commit --no-verify
```

---

### 3. Clean Test Runner

**Purpose:** Runs Go tests while filtering out GORM warning messages.

**Unix/Linux/Mac/Git Bash:**
```bash
bash scripts/test-clean.sh [test-flags] [packages]
```

**Windows Command Prompt:**
```cmd
scripts\test-clean.bat [test-flags] [packages]
```

**Problem:**
GORM outputs numerous warnings about BeforeCreate/BeforeUpdate/BeforeDelete hook interfaces when running tests with SQLite. These warnings are harmless (the code works correctly with PostgreSQL in production) but clutter test output.

**Solution:**
The clean test runner filters out these warnings while preserving actual test results.

**Examples:**
```bash
# Run all tests
bash scripts/test-clean.sh ./...

# Run with verbose output
bash scripts/test-clean.sh -v ./...

# Run specific package
bash scripts/test-clean.sh ./tests/services/...

# Run with coverage
bash scripts/test-clean.sh -cover ./...

# Run only short tests
bash scripts/test-clean.sh -short ./...
```

**What it filters:**
- ✅ GORM hook mismatch warnings (`don't match BeforeCreateInterface`, etc.)
- ✅ Associated timestamp and file path lines
- ✅ ANSI color code artifacts
- ❌ **Actual test failures and errors are NOT filtered** - they will still show up

**Output Comparison:**

Without script (verbose warnings):
```
=== RUN   TestProductService_CreateProduct_Success

2025/11/07 01:15:15 [warn] Model don't match BeforeCreateInterface...
2025/11/07 01:15:15 [warn] Model don't match BeforeUpdateInterface...
(hundreds of warning lines)

--- PASS: TestProductService_CreateProduct_Success (0.12s)
```

With script (clean output):
```
=== RUN   TestProductService_CreateProduct_Success
--- PASS: TestProductService_CreateProduct_Success (0.12s)
=== RUN   TestProductService_GetProduct_Success
--- PASS: TestProductService_GetProduct_Success (0.11s)
```

---

## Quick Setup Guide

### First Time Setup (New Developer)

1. **Install development tools:**
   ```bash
   bash scripts/setup-dev-tools.sh
   ```

2. **Install Git hooks:**
   ```bash
   bash scripts/install-hooks.sh
   ```

3. **Verify setup:**
   ```bash
   # Check tools are installed
   which goimports golangci-lint gosec

   # Check hook is installed
   ls -la .git/hooks/pre-commit
   ```

4. **Test the hook:**
   ```bash
   # Make a small change and commit
   git add .
   git commit -m "test: verify pre-commit hook"
   # All 10 checks should run
   ```

---

### After Pulling Hook Updates

When `.git-hooks/pre-commit` is updated in the repository:

```bash
# Reinstall the hook to get the latest version
bash scripts/install-hooks.sh
```

---

## Troubleshooting

### Scripts Not Executable (Unix)

If you get "Permission denied":
```bash
chmod +x scripts/*.sh
```

### GOBIN Not in PATH

If tools install but aren't found:

**Unix/Linux/Mac (bash/zsh):**
Add to `~/.bashrc` or `~/.zshrc`:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Then reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

**Windows:**
1. Open "Environment Variables" in System Settings
2. Edit the "Path" variable
3. Add: `%USERPROFILE%\go\bin`
4. Restart terminal

### Installation Fails

If tool installation fails:

1. **Check Go installation:**
   ```bash
   go version
   ```

2. **Check internet connection:**
   Tools are downloaded from GitHub and golang.org

3. **Manually install:**
   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/securego/gosec/v2/cmd/gosec@latest
   ```

4. **Use alternative methods:**
   - golangci-lint: https://golangci-lint.run/usage/install/
   - Others typically only available via `go install`

### Test Script Issues

**On Windows:** If `test-clean.bat` doesn't work, use PowerShell or Git Bash instead:
```bash
bash scripts/test-clean.sh ./...
```

**On Unix:** If script isn't executable:
```bash
chmod +x scripts/test-clean.sh
bash scripts/test-clean.sh ./...
```

---

## Script Maintenance

### Adding New Scripts

When adding new scripts to this directory:

1. **Create the script:**
   - Unix: `.sh` extension with `#!/bin/bash` shebang
   - Windows: `.bat` extension with `@echo off`

2. **Make it executable (Unix):**
   ```bash
   chmod +x scripts/your-script.sh
   ```

3. **Document it:**
   - Add a comment block at the top of the script
   - Update this README with usage instructions

4. **Test on both platforms:**
   - Test shell script on Unix/Linux/Mac/Git Bash
   - Test batch script on Windows Command Prompt

---

## Related Documentation

- **Git Hooks Documentation:** `.git-hooks/README.md`
- **golangci-lint Configuration:** `.golangci.yml`
- **Test Utilities:** `tests/testutils/`
- **Project Setup:** Main `README.md`

---

## CI/CD Integration

These scripts are designed for local development. For CI/CD:

**GitHub Actions:**
```yaml
name: Quality Checks

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install tools
        run: bash scripts/setup-dev-tools.sh

      - name: Run linting
        run: golangci-lint run --timeout=5m

      - name: Run security scan
        run: gosec -quiet ./...

      - name: Run tests
        run: bash scripts/test-clean.sh -v -cover ./...
```

**GitLab CI:**
```yaml
stages:
  - test

quality-checks:
  stage: test
  image: golang:1.21
  script:
    - bash scripts/setup-dev-tools.sh
    - golangci-lint run --timeout=5m
    - gosec -quiet ./...
    - bash scripts/test-clean.sh -v ./...
```

**Docker:**
```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

# Install tools
RUN bash scripts/setup-dev-tools.sh

# Run checks
RUN golangci-lint run --timeout=5m
RUN gosec ./...
RUN go test -v ./...
```

---

## Alternative: Direct Command Usage

You can also run tools directly without scripts:

```bash
# Format code
goimports -w .

# Run linter
golangci-lint run --timeout=2m

# Security scan
gosec -quiet ./...

# Regular tests
go test -v ./...

# Tests with coverage
go test -v -cover ./...
```

The scripts just provide convenience and consistent output formatting.
