#!/bin/bash
# Development Tools Setup Script
# Installs all required tools for enhanced pre-commit hooks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo ""
echo -e "${BLUE}🔧 Kisanlink ERP - Development Tools Setup${NC}"
echo "================================================"
echo ""
echo "This script will install the following tools:"
echo "  1. goimports  - Code formatting & import organization"
echo "  2. golangci-lint - Fast linters runner"
echo "  3. gosec - Security vulnerability scanner"
echo ""

INSTALLED_COUNT=0
SKIPPED_COUNT=0
FAILED_COUNT=0

# Function to check if a command exists
command_exists() {
    command -v "$1" &> /dev/null
}

# Function to get Go bin directory
get_gobin() {
    if [ -n "$GOBIN" ]; then
        echo "$GOBIN"
    elif [ -n "$GOPATH" ]; then
        echo "$GOPATH/bin"
    else
        echo "$HOME/go/bin"
    fi
}

GOBIN_DIR=$(get_gobin)

# Check if GOBIN is in PATH
if [[ ":$PATH:" != *":$GOBIN_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  Warning: $GOBIN_DIR is not in your PATH${NC}"
    echo -e "${YELLOW}   Add this to your ~/.bashrc or ~/.zshrc:${NC}"
    echo -e "${YELLOW}   export PATH=\$PATH:$GOBIN_DIR${NC}"
    echo ""
fi

# ============================================
# 1. Install goimports
# ============================================
echo -n "1️⃣  Installing goimports... "

if command_exists goimports; then
    echo -e "${GREEN}✓${NC} (already installed)"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
else
    if go install golang.org/x/tools/cmd/goimports@latest; then
        if command_exists goimports; then
            echo -e "${GREEN}✓${NC} (installed successfully)"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        else
            echo -e "${YELLOW}⚠${NC} (installed but not found in PATH)"
            echo -e "${YELLOW}   Installed to: $GOBIN_DIR/goimports${NC}"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        fi
    else
        echo -e "${RED}✗${NC} (installation failed)"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
fi

# ============================================
# 2. Install golangci-lint
# ============================================
echo -n "2️⃣  Installing golangci-lint... "

if command_exists golangci-lint; then
    CURRENT_VERSION=$(golangci-lint --version 2>/dev/null | grep -oP 'version \K[0-9.]+' || echo "unknown")
    echo -e "${GREEN}✓${NC} (already installed - v$CURRENT_VERSION)"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
else
    # Try the official install script method
    if curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$GOBIN_DIR" &>/dev/null; then
        if command_exists golangci-lint; then
            VERSION=$(golangci-lint --version 2>/dev/null | grep -oP 'version \K[0-9.]+' || echo "unknown")
            echo -e "${GREEN}✓${NC} (installed v$VERSION)"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        else
            echo -e "${YELLOW}⚠${NC} (installed but not found in PATH)"
            echo -e "${YELLOW}   Installed to: $GOBIN_DIR/golangci-lint${NC}"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        fi
    else
        # Fallback to go install
        echo -n "(trying alternative method)... "
        if go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; then
            if command_exists golangci-lint; then
                echo -e "${GREEN}✓${NC} (installed successfully)"
                INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
            else
                echo -e "${YELLOW}⚠${NC} (installed but not found in PATH)"
                echo -e "${YELLOW}   Installed to: $GOBIN_DIR/golangci-lint${NC}"
                INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
            fi
        else
            echo -e "${RED}✗${NC} (installation failed)"
            echo -e "${RED}   Manual installation required${NC}"
            echo -e "${YELLOW}   Visit: https://golangci-lint.run/usage/install/${NC}"
            FAILED_COUNT=$((FAILED_COUNT + 1))
        fi
    fi
fi

# ============================================
# 3. Install gosec
# ============================================
echo -n "3️⃣  Installing gosec... "

if command_exists gosec; then
    CURRENT_VERSION=$(gosec -version 2>/dev/null | grep -oP 'Version: \K[0-9.]+' || echo "unknown")
    echo -e "${GREEN}✓${NC} (already installed - v$CURRENT_VERSION)"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
else
    if go install github.com/securego/gosec/v2/cmd/gosec@latest; then
        if command_exists gosec; then
            VERSION=$(gosec -version 2>/dev/null | grep -oP 'Version: \K[0-9.]+' || echo "unknown")
            echo -e "${GREEN}✓${NC} (installed v$VERSION)"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        else
            echo -e "${YELLOW}⚠${NC} (installed but not found in PATH)"
            echo -e "${YELLOW}   Installed to: $GOBIN_DIR/gosec${NC}"
            INSTALLED_COUNT=$((INSTALLED_COUNT + 1))
        fi
    else
        echo -e "${RED}✗${NC} (installation failed)"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
fi

# ============================================
# SUMMARY
# ============================================
echo ""
echo "================================================"
echo -e "${BLUE}📊 Installation Summary${NC}"
echo ""
echo -e "  ${GREEN}✅ Installed: $INSTALLED_COUNT${NC}"
echo -e "  ${YELLOW}⏭️  Skipped:   $SKIPPED_COUNT${NC}"
echo -e "  ${RED}❌ Failed:    $FAILED_COUNT${NC}"
echo ""

if [ $FAILED_COUNT -eq 0 ]; then
    echo -e "${GREEN}✅ All tools are ready!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Install Git hooks: bash scripts/install-hooks.sh"
    echo "  2. Test pre-commit: git add . && git commit -m \"test\""
    echo ""
else
    echo -e "${YELLOW}⚠️  Some tools failed to install${NC}"
    echo ""
    echo "You can still use the pre-commit hook, but some checks will be skipped."
    echo ""
    echo "To manually install failed tools, see:"
    echo "  - goimports:      go install golang.org/x/tools/cmd/goimports@latest"
    echo "  - golangci-lint:  https://golangci-lint.run/usage/install/"
    echo "  - gosec:          go install github.com/securego/gosec/v2/cmd/gosec@latest"
    echo ""
fi

# ============================================
# VERIFICATION
# ============================================
if [ $INSTALLED_COUNT -gt 0 ] || [ $SKIPPED_COUNT -gt 0 ]; then
    echo "================================================"
    echo -e "${BLUE}🔍 Verifying installations...${NC}"
    echo ""

    if command_exists goimports; then
        echo -e "  goimports:      ${GREEN}✓${NC} $(command -v goimports)"
    else
        echo -e "  goimports:      ${RED}✗ Not found in PATH${NC}"
    fi

    if command_exists golangci-lint; then
        echo -e "  golangci-lint:  ${GREEN}✓${NC} $(command -v golangci-lint)"
    else
        echo -e "  golangci-lint:  ${RED}✗ Not found in PATH${NC}"
    fi

    if command_exists gosec; then
        echo -e "  gosec:          ${GREEN}✓${NC} $(command -v gosec)"
    else
        echo -e "  gosec:          ${RED}✗ Not found in PATH${NC}"
    fi

    echo ""
fi

echo "================================================"
echo ""

exit 0
