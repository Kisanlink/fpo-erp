@echo off
REM Install Git hooks from .git-hooks directory

echo.
echo Installing Git hooks...
echo.

REM Check if .git-hooks directory exists
if not exist ".git-hooks" (
    echo Error: .git-hooks directory not found
    echo Please run this script from the repository root
    exit /b 1
)

REM Check if .git directory exists
if not exist ".git" (
    echo Error: .git directory not found
    echo Please run this script from the repository root
    exit /b 1
)

REM Copy pre-commit hook
if exist ".git-hooks\pre-commit" (
    copy /Y ".git-hooks\pre-commit" ".git\hooks\pre-commit" >nul
    echo Installed pre-commit hook
) else (
    echo Warning: .git-hooks\pre-commit not found
)

echo.
echo Git hooks installation complete!
echo.
echo The following hooks are now active:
echo   - pre-commit: Runs tests before every commit (non-blocking)
echo.
echo To skip hooks when committing, use: git commit --no-verify
echo.
