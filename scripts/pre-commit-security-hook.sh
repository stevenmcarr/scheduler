#!/bin/bash
# Pre-commit hook to prevent committing sensitive information
# Place this file in .git/hooks/pre-commit and make it executable

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
error() {
    echo -e "${RED}[SECURITY ERROR]${NC} $1" >&2
}

warning() {
    echo -e "${YELLOW}[SECURITY WARNING]${NC} $1" >&2
}

# Check for sensitive files
sensitive_files=$(git diff --cached --name-only | grep -E '\.(env|key|pem|crt|p12|pfx)$|password|secret|credential')

if [[ -n "$sensitive_files" ]]; then
    error "Attempting to commit sensitive files:"
    echo "$sensitive_files"
    error "These files may contain sensitive information and should not be committed."
    exit 1
fi

# Check for hardcoded passwords in staged content
if git diff --cached | grep -i -E 'password\s*=\s*["\'][^"\']+["\']|secret\s*=\s*["\'][^"\']+["\']|token\s*=\s*["\'][^"\']+["\']'; then
    error "Found potential hardcoded credentials in staged changes."
    error "Please use environment variables instead of hardcoding sensitive data."
    exit 1
fi

# Check for .env files (should be in .gitignore)
if git diff --cached --name-only | grep -E '^\.env$|^\.env\.'; then
    error "Attempting to commit .env file(s)."
    error "Environment files should never be committed. Add them to .gitignore."
    exit 1
fi

# Warning for database dumps
if git diff --cached --name-only | grep -E '\.(sql|dump|backup)$'; then
    warning "You are committing database files. Ensure they don't contain sensitive data."
fi

echo "✅ Security pre-commit checks passed."
