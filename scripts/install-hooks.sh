#!/bin/bash
# Install Git hooks from .git-hooks directory

echo "📦 Installing Git hooks..."
echo ""

# Check if .git-hooks directory exists
if [ ! -d ".git-hooks" ]; then
    echo "❌ Error: .git-hooks directory not found"
    echo "   Please run this script from the repository root"
    exit 1
fi

# Check if .git directory exists
if [ ! -d ".git" ]; then
    echo "❌ Error: .git directory not found"
    echo "   Please run this script from the repository root"
    exit 1
fi

# Copy pre-commit hook
if [ -f ".git-hooks/pre-commit" ]; then
    cp .git-hooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "✅ Installed pre-commit hook"
else
    echo "⚠️  Warning: .git-hooks/pre-commit not found"
fi

echo ""
echo "✨ Git hooks installation complete!"
echo ""
echo "The following hooks are now active:"
echo "  - pre-commit: Runs tests before every commit (non-blocking)"
echo ""
echo "To skip hooks when committing, use: git commit --no-verify"
echo ""
