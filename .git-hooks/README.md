# Git Hooks

This directory contains Git hooks that can be installed to automate tasks during Git operations.

## Available Hooks

### pre-commit (Enhanced)
Comprehensive quality checks before every `git commit` operation.

**10 Quality Checks:**
1. ✅ **Branch Protection (BLOCKING)** - Prevents direct commits to main/master
2. ✅ **Merge Conflict Markers (BLOCKING)** - Detects unresolved merge conflicts
3. ✅ **Code Formatting (AUTO-FIX)** - Runs gofmt/goimports and auto-formats code
4. ✅ **Go Module Tidy (AUTO-FIX)** - Automatically tidies go.mod and go.sum
5. ✅ **Build Check (WARNING)** - Verifies project builds successfully
6. ✅ **Linting (WARNING/BLOCKING)** - Runs golangci-lint for code quality
7. ✅ **Tests (WARNING)** - Runs all tests with clean output
8. ✅ **Security Scan (WARNING/BLOCKING)** - Runs gosec for vulnerability detection
9. ✅ **Debug Code Detection (WARNING)** - Finds potential debug code (fmt.Println, panic, etc.)
10. ✅ **Sensitive Data (WARNING)** - Detects potential secrets and credentials
11. 📝 **TODO/FIXME Reporting (INFO)** - Lists all TODO/FIXME comments

**Behavior:**
- 🚫 **2 Blocking Checks** - Will prevent commit if failed
- 🛠️ **2 Auto-Fix Checks** - Automatically fixes issues and restages files
- ⚠️ **6 Warning Checks** - Shows warnings but allows commit
- 📊 **Summary Report** - Shows counts and execution time
- ⚙️ **Configurable** - Use environment variables to skip/block checks

**Configuration Options:**
```bash
# Skip specific checks
SKIP_FORMAT=1 git commit -m "message"      # Skip code formatting
SKIP_LINT=1 git commit -m "message"        # Skip linting
SKIP_SECURITY=1 git commit -m "message"    # Skip security scan
SKIP_BUILD=1 git commit -m "message"       # Skip build check
SKIP_TESTS=1 git commit -m "message"       # Skip tests

# Make warnings blocking
BLOCK_ON_LINT=1 git commit -m "message"    # Block commit if linting fails
BLOCK_ON_SECURITY=1 git commit -m "message" # Block commit if security issues found

# Skip all checks
git commit --no-verify -m "message"
```

**Example Output (all checks pass):**
```
🔍 Running pre-commit checks...
================================================
1️⃣  Checking branch... ✓ (feat/feature-name)
2️⃣  Checking for merge conflicts... ✓
3️⃣  Formatting code... ✓
4️⃣  Tidying Go modules... ✓
5️⃣  Building project... ✓
6️⃣  Running linter... ✓
7️⃣  Running tests... ✓ (27 packages passed)
8️⃣  Security scanning... ✓
9️⃣  Checking for debug code... ✓
🔟 Checking for sensitive data... ✓
📝 Checking TODOs/FIXMEs... ✓ (none)

================================================
✅ ALL CHECKS PASSED
⏱️  Time: 15s

Proceeding with commit...
================================================
```

**Example Output (with warnings):**
```
🔍 Running pre-commit checks...
================================================
1️⃣  Checking branch... ✓ (feat/feature-name)
2️⃣  Checking for merge conflicts... ✓
3️⃣  Formatting code... ✓ (3 files formatted)
4️⃣  Tidying Go modules... ✓ (go.mod updated)
5️⃣  Building project... ✓
6️⃣  Running linter... ✗
⚠️  Linting issues found:
   internal/services/product.go:45: errcheck: Error return value is not checked
   internal/handlers/user.go:123: unused: unused variable 'result'
   ... and 3 more issues
7️⃣  Running tests... ✗
⚠️  Test failures:
   --- FAIL: TestProductService_SearchProducts (0.10s)
8️⃣  Security scanning... ✓
9️⃣  Checking for debug code... ⚠
⚠️  Found potential debug code:
   internal/services/product.go:67:fmt.Println("Debug: processing product")
🔟 Checking for sensitive data... ✓
📝 Checking TODOs/FIXMEs... ℹ (5 found)
   internal/handlers/product.go:34:// TODO: Add pagination support
   internal/services/user.go:89:// FIXME: Handle edge case
   ... and 3 more

================================================
⚠️  COMMITTING WITH 3 WARNING(S)
   Please review and fix warnings when possible
⏱️  Time: 18s

Proceeding with commit...
================================================
```

## Installation

### Prerequisites

The enhanced pre-commit hook requires these development tools:
- `goimports` - Code formatting and import organization
- `golangci-lint` - Comprehensive linting
- `gosec` - Security vulnerability scanning

**Quick Install (Recommended):**

Unix/Linux/Mac/Git Bash:
```bash
bash scripts/setup-dev-tools.sh
```

Windows Command Prompt:
```cmd
scripts\setup-dev-tools.bat
```

**Manual Install:**
```bash
# goimports
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
# Or visit: https://golangci-lint.run/usage/install/

# gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

### One-Time Setup

After installing the tools, install the Git hooks:

**Unix/Linux/Mac/Git Bash:**
```bash
bash scripts/install-hooks.sh
```

**Windows Command Prompt:**
```cmd
scripts\install-hooks.bat
```

This copies the hooks from `.git-hooks/` to `.git/hooks/` and makes them executable.

**Note:** The pre-commit hook will still run even if tools are missing - it will skip checks for unavailable tools and show installation instructions.

## Usage

### Normal Commits
Once installed, hooks run automatically with all checks:
```bash
git add .
git commit -m "Your commit message"
# All 10 checks run automatically before commit
```

### Skip Specific Checks
Use environment variables to skip individual checks:
```bash
# Skip formatting (use if you haven't installed goimports yet)
SKIP_FORMAT=1 git commit -m "message"

# Skip linting (faster commits during development)
SKIP_LINT=1 git commit -m "message"

# Skip security scan (faster commits)
SKIP_SECURITY=1 git commit -m "message"

# Skip build check
SKIP_BUILD=1 git commit -m "message"

# Skip tests (fastest option)
SKIP_TESTS=1 git commit -m "message"

# Combine multiple skips
SKIP_LINT=1 SKIP_SECURITY=1 git commit -m "message"
```

### Make Checks Blocking
Convert warnings to blocking errors:
```bash
# Block commit if linting issues found
BLOCK_ON_LINT=1 git commit -m "message"

# Block commit if security issues found
BLOCK_ON_SECURITY=1 git commit -m "message"

# Both
BLOCK_ON_LINT=1 BLOCK_ON_SECURITY=1 git commit -m "message"
```

### Skip All Hooks
To bypass all checks temporarily:
```bash
git commit --no-verify -m "Quick fix"
```

## Customization

### Permanent Configuration
Edit `.git-hooks/pre-commit` to change default behavior:

**Skip checks by default:**
```bash
# Near the top of the file, change from:
SKIP_LINT=${SKIP_LINT:-0}

# To:
SKIP_LINT=${SKIP_LINT:-1}  # Skips linting by default
```

**Make checks blocking by default:**
```bash
# Change from:
BLOCK_ON_LINT=${BLOCK_ON_LINT:-0}

# To:
BLOCK_ON_LINT=${BLOCK_ON_LINT:-1}  # Blocks on lint errors
```

After editing, reinstall the hook:
```bash
bash scripts/install-hooks.sh
```

### Customize Linting Rules
Edit `.golangci.yml` in the project root to:
- Enable/disable specific linters
- Adjust timeout values
- Add custom exclusions
- Configure linter-specific settings

See [golangci-lint documentation](https://golangci-lint.run/usage/configuration/) for details.

### Customize Test Execution
Edit the test command in `.git-hooks/pre-commit`:

```bash
# Line ~170: Change test command

# Run only fast tests
TEST_OUTPUT=$(go test -short ./... 2>&1)

# Run only service tests
TEST_OUTPUT=$(bash scripts/test-clean.sh ./tests/services/... 2>&1)

# Run with race detector
TEST_OUTPUT=$(go test -race ./... 2>&1)
```

## Troubleshooting

### Hook Not Running
1. Check if hook is installed: `ls -la .git/hooks/pre-commit`
2. Reinstall: `bash scripts/install-hooks.sh`
3. Verify executable (Unix): `chmod +x .git/hooks/pre-commit`
4. Check for errors: Try committing to see error messages

### Tools Not Found
If you see "not found" warnings for goimports, golangci-lint, or gosec:

1. **Install tools**: Run `bash scripts/setup-dev-tools.sh`
2. **Check PATH**: Ensure `$GOPATH/bin` or `$GOBIN` is in your PATH
3. **Verify installation**: Run `which goimports golangci-lint gosec`
4. **Skip checks temporarily**: Use `SKIP_FORMAT=1 SKIP_LINT=1 SKIP_SECURITY=1 git commit -m "message"`

### Pre-Commit Takes Too Long
If checks take more than 30 seconds:

1. **Skip slow checks**:
   ```bash
   SKIP_TESTS=1 SKIP_LINT=1 git commit -m "message"
   ```

2. **Run only fast tests**: Edit `.git-hooks/pre-commit` line ~170:
   ```bash
   TEST_OUTPUT=$(go test -short ./... 2>&1)
   ```

3. **Use faster linting**: In `.golangci.yml`, add `fast: true` (already enabled)

4. **Skip temporarily**:
   ```bash
   git commit --no-verify -m "Quick fix"
   ```

### Linting Errors Block Commit
If linting errors prevent commits (when `BLOCK_ON_LINT=1`):

1. **Fix the issues**: Address the reported linting problems
2. **Temporarily allow**: Use `BLOCK_ON_LINT=0 git commit -m "message"`
3. **Disable rule**: Add exclusion to `.golangci.yml`
4. **Skip entirely**: Use `SKIP_LINT=1 git commit -m "message"`

### Security Scan False Positives
If gosec reports false positives:

1. **Add comment in code**:
   ```go
   // #nosec G104 -- Error intentionally ignored
   _ = file.Close()
   ```

2. **Skip security check**:
   ```bash
   SKIP_SECURITY=1 git commit -m "message"
   ```

3. **Configure gosec**: Edit `.golangci.yml` gosec section to exclude specific rules

### Debug Code Warnings
If you need to keep debug code temporarily:

The check is informational only - commits will proceed. To suppress:
- Remove debug statements (`fmt.Println`, etc.)
- Or accept the warning and commit anyway

## Why .git-hooks Directory?

Git hooks in `.git/hooks/` cannot be committed to the repository (`.git/` is gitignored).

By keeping hooks in `.git-hooks/` (committed to repo) and providing install scripts, we ensure:
- ✅ Hooks are version-controlled
- ✅ All team members can easily install them
- ✅ Hook updates are distributed via git pull
- ✅ Easy to modify and test hooks before installing
