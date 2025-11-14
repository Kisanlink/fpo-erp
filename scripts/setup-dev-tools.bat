@echo off
REM Development Tools Setup Script for Windows
REM Installs all required tools for enhanced pre-commit hooks

setlocal enabledelayedexpansion

echo.
echo ==================================================
echo    Kisanlink ERP - Development Tools Setup
echo ==================================================
echo.
echo This script will install the following tools:
echo   1. goimports     - Code formatting ^& import organization
echo   2. golangci-lint - Fast linters runner
echo   3. gosec         - Security vulnerability scanner
echo.

set INSTALLED_COUNT=0
set SKIPPED_COUNT=0
set FAILED_COUNT=0

REM Get GOBIN directory
if defined GOBIN (
    set "GOBIN_DIR=%GOBIN%"
) else if defined GOPATH (
    set "GOBIN_DIR=%GOPATH%\bin"
) else (
    set "GOBIN_DIR=%USERPROFILE%\go\bin"
)

REM Check if GOBIN is in PATH
echo %PATH% | findstr /C:"%GOBIN_DIR%" >nul
if errorlevel 1 (
    echo WARNING: %GOBIN_DIR% is not in your PATH
    echo    Add it to your system environment variables
    echo.
)

REM ============================================
REM 1. Install goimports
REM ============================================
echo Installing goimports...

where goimports >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] goimports already installed
    set /a SKIPPED_COUNT+=1
) else (
    echo    Installing goimports...
    go install golang.org/x/tools/cmd/goimports@latest >nul 2>&1
    if %errorlevel% equ 0 (
        where goimports >nul 2>&1
        if %errorlevel% equ 0 (
            echo [OK] goimports installed successfully
            set /a INSTALLED_COUNT+=1
        ) else (
            echo [WARNING] goimports installed but not found in PATH
            echo    Installed to: %GOBIN_DIR%\goimports.exe
            set /a INSTALLED_COUNT+=1
        )
    ) else (
        echo [ERROR] goimports installation failed
        set /a FAILED_COUNT+=1
    )
)
echo.

REM ============================================
REM 2. Install golangci-lint
REM ============================================
echo Installing golangci-lint...

where golangci-lint >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=2" %%v in ('golangci-lint --version 2^>nul ^| findstr /C:"version"') do (
        echo [OK] golangci-lint already installed - %%v
    )
    set /a SKIPPED_COUNT+=1
) else (
    echo    Installing golangci-lint via go install...
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest >nul 2>&1
    if %errorlevel% equ 0 (
        where golangci-lint >nul 2>&1
        if %errorlevel% equ 0 (
            for /f "tokens=2" %%v in ('golangci-lint --version 2^>nul ^| findstr /C:"version"') do (
                echo [OK] golangci-lint installed - %%v
            )
            set /a INSTALLED_COUNT+=1
        ) else (
            echo [WARNING] golangci-lint installed but not found in PATH
            echo    Installed to: %GOBIN_DIR%\golangci-lint.exe
            set /a INSTALLED_COUNT+=1
        )
    ) else (
        echo [ERROR] golangci-lint installation failed
        echo    Manual installation required
        echo    Visit: https://golangci-lint.run/usage/install/
        set /a FAILED_COUNT+=1
    )
)
echo.

REM ============================================
REM 3. Install gosec
REM ============================================
echo Installing gosec...

where gosec >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=2" %%v in ('gosec -version 2^>nul ^| findstr /C:"Version"') do (
        echo [OK] gosec already installed - %%v
    )
    set /a SKIPPED_COUNT+=1
) else (
    echo    Installing gosec...
    go install github.com/securego/gosec/v2/cmd/gosec@latest >nul 2>&1
    if %errorlevel% equ 0 (
        where gosec >nul 2>&1
        if %errorlevel% equ 0 (
            for /f "tokens=2" %%v in ('gosec -version 2^>nul ^| findstr /C:"Version"') do (
                echo [OK] gosec installed - %%v
            )
            set /a INSTALLED_COUNT+=1
        ) else (
            echo [WARNING] gosec installed but not found in PATH
            echo    Installed to: %GOBIN_DIR%\gosec.exe
            set /a INSTALLED_COUNT+=1
        )
    ) else (
        echo [ERROR] gosec installation failed
        set /a FAILED_COUNT+=1
    )
)
echo.

REM ============================================
REM SUMMARY
REM ============================================
echo ==================================================
echo Installation Summary
echo.
echo   Installed: %INSTALLED_COUNT%
echo   Skipped:   %SKIPPED_COUNT%
echo   Failed:    %FAILED_COUNT%
echo.

if %FAILED_COUNT% equ 0 (
    echo [SUCCESS] All tools are ready!
    echo.
    echo Next steps:
    echo   1. Install Git hooks: scripts\install-hooks.bat
    echo   2. Test pre-commit: git add . ^&^& git commit -m "test"
    echo.
) else (
    echo [WARNING] Some tools failed to install
    echo.
    echo You can still use the pre-commit hook, but some checks will be skipped.
    echo.
    echo To manually install failed tools, see:
    echo   - goimports:      go install golang.org/x/tools/cmd/goimports@latest
    echo   - golangci-lint:  https://golangci-lint.run/usage/install/
    echo   - gosec:          go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo.
)

REM ============================================
REM VERIFICATION
REM ============================================
if %INSTALLED_COUNT% gtr 0 (
    echo ==================================================
    echo Verifying installations...
    echo.

    where goimports >nul 2>&1
    if %errorlevel% equ 0 (
        for /f "delims=" %%p in ('where goimports') do (
            echo   goimports:      [OK] %%p
        )
    ) else (
        echo   goimports:      [ERROR] Not found in PATH
    )

    where golangci-lint >nul 2>&1
    if %errorlevel% equ 0 (
        for /f "delims=" %%p in ('where golangci-lint') do (
            echo   golangci-lint:  [OK] %%p
        )
    ) else (
        echo   golangci-lint:  [ERROR] Not found in PATH
    )

    where gosec >nul 2>&1
    if %errorlevel% equ 0 (
        for /f "delims=" %%p in ('where gosec') do (
            echo   gosec:          [OK] %%p
        )
    ) else (
        echo   gosec:          [ERROR] Not found in PATH
    )

    echo.
)

echo ==================================================
echo.

pause
exit /b 0
