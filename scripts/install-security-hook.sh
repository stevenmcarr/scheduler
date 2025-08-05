#!/bin/bash
# Install the security pre-commit hook

set -e

HOOK_SOURCE="$(pwd)/scripts/pre-commit-security-hook.sh"
HOOK_DEST=".git/hooks/pre-commit"

echo "Installing security pre-commit hook..."

# Check if we're in a git repository
if [[ ! -d ".git" ]]; then
    echo "Error: Not in a git repository root directory."
    exit 1
fi

# Check if hook source exists
if [[ ! -f "$HOOK_SOURCE" ]]; then
    echo "Error: Hook source file not found: $HOOK_SOURCE"
    exit 1
fi

# Backup existing hook if it exists
if [[ -f "$HOOK_DEST" ]]; then
    echo "Backing up existing pre-commit hook..."
    cp "$HOOK_DEST" "$HOOK_DEST.backup.$(date +%Y%m%d_%H%M%S)"
fi

# Copy and make executable
cp "$HOOK_SOURCE" "$HOOK_DEST"
chmod +x "$HOOK_DEST"

echo "✅ Security pre-commit hook installed successfully!"
echo "The hook will now prevent committing sensitive files and data."
echo ""
echo "To temporarily bypass the hook (not recommended):"
echo "  git commit --no-verify"
echo ""
echo "To uninstall the hook:"
echo "  rm .git/hooks/pre-commit"
