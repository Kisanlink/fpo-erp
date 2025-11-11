# Pre-commit Hooks Setup Guide

This project uses [pre-commit](https://pre-commit.com/) to ensure code quality and consistency across all commits.

## Prerequisites

### Install pre-commit

#### macOS
```bash
brew install pre-commit
```

#### Linux
```bash
pip install pre-commit
# or
apt-get install pre-commit  # Ubuntu/Debian
```

#### Windows
```bash
pip install pre-commit
# or using scoop
scoop install pre-commit
```

## Installation

Once pre-commit is installed, run:

```bash
make install-hooks
```

This will:
1. Install pre-commit hooks into `.git/hooks/`
2. Configure both pre-commit and commit-msg hooks
3. Verify the installation

## Hooks Configured

### Go Quality Checks
- **go-fmt**: Formats Go code using `gofmt`
- **go-imports**: Organizes imports using `goimports`
- **go-mod-tidy**: Cleans up `go.mod` and `go.sum`
- **go-unit-tests**: Runs unit tests with `-short` and `-race` flags
- **golangci-lint**: Comprehensive Go linting (5 minute timeout)
- **gosec**: Security scanning for Go code

### General Quality Checks
- **check-added-large-files**: Prevents files larger than 1MB
- **check-case-conflict**: Detects file name case conflicts
- **check-merge-conflict**: Prevents commits with merge conflict markers
- **check-yaml**: Validates YAML files
- **check-json**: Validates JSON files
- **end-of-file-fixer**: Ensures files end with newline
- **trailing-whitespace**: Removes trailing whitespace
- **mixed-line-ending**: Enforces LF line endings
- **detect-private-key**: Prevents committing private keys
- **no-commit-to-branch**: Prevents direct commits to main/master

### Documentation
- **swagger-validate**: Regenerates Swagger documentation

### Commit Message Format
- **conventional-pre-commit**: Enforces conventional commit format
  - Examples: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`

## Usage

### Automatic Execution
Hooks run automatically on:
- `git commit` (pre-commit hooks)
- Commit message validation (commit-msg hook)

### Manual Execution
Run hooks on all files:
```bash
make test-hooks
```

Run hooks on staged files only:
```bash
pre-commit run
```

### Bypassing Hooks (Not Recommended)
If absolutely necessary (e.g., work in progress):
```bash
git commit --no-verify -m "wip: message"
```

**Note**: This is discouraged. Fix issues instead of bypassing hooks.

## Updating Hooks
To update hooks to their latest versions:
```bash
make update-hooks
```

## Uninstalling Hooks
To remove pre-commit hooks:
```bash
make uninstall-hooks
```

## Cross-Platform Compatibility

The hooks are designed to work on:
- **macOS**: Native support via Homebrew
- **Linux**: Native support via pip or package managers
- **Windows**: Works with Git Bash, WSL, or native Python installation

### Windows-Specific Notes
- Install Git Bash or use WSL for best compatibility
- Ensure Python is in your PATH
- Use forward slashes or double backslashes in paths

## Troubleshooting

### Hook fails with "command not found"
Ensure required tools are installed:
```bash
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Hooks are too slow
Consider:
1. Running specific hooks: `pre-commit run <hook-id>`
2. Skipping expensive checks locally (not recommended for CI/CD)

### Pre-commit not found
Verify installation:
```bash
which pre-commit  # macOS/Linux
where pre-commit  # Windows
```

If not found, reinstall using the instructions above.

## CI/CD Integration

In GitHub Actions or other CI/CD:
```yaml
- name: Run pre-commit hooks
  run: |
    pip install pre-commit
    pre-commit run --all-files
```

## Additional Resources
- [pre-commit documentation](https://pre-commit.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [golangci-lint](https://golangci-lint.run/)
