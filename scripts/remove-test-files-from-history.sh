#!/bin/bash
# Script to completely remove test_ files from git history
# WARNING: This rewrites git history and requires force push

echo "🚨 WARNING: This will rewrite git history!"
echo "This operation will remove test_ files completely from all commits."
echo "You will need to force push after this operation."
echo ""
read -p "Are you sure you want to continue? (yes/no): " confirm

if [[ "$confirm" != "yes" ]]; then
    echo "Operation cancelled."
    exit 0
fi

echo "Removing test_ files from git history..."

# Use git filter-branch to remove test_ files from history
git filter-branch --force --index-filter \
    'git rm --cached --ignore-unmatch scripts/test_mysql_connection.sh scripts/test_mysql_fix.sh scripts/test_password_prompt.sh scripts/test_stderr_fix.sh' \
    --prune-empty --tag-name-filter cat -- --all

# Clean up the backup refs
git for-each-ref --format="%(refname)" refs/original/ | xargs -n 1 git update-ref -d

# Expire reflog and garbage collect
git reflog expire --expire=now --all
git gc --prune=now --aggressive

echo "✅ Test files removed from git history."
echo ""
echo "Next steps:"
echo "1. Verify the files are gone: git log --name-only | grep test_"
echo "2. Force push to remote: git push --force-with-lease origin main"
echo "3. Team members will need to re-clone or reset their repos"
