# Security Guidelines for WMU Course Scheduler

## 🔐 Password and Credential Management

### NEVER Do This:
❌ Hardcode passwords in source code  
❌ Commit `.env` files to git  
❌ Share passwords in chat or email  
❌ Use weak or default passwords  
❌ Store passwords in plain text files  

### ALWAYS Do This:
✅ Use environment variables for sensitive data  
✅ Use strong, unique passwords  
✅ Use `.env.example` templates with placeholder values  
✅ Add sensitive files to `.gitignore`  
✅ Use secure password managers  
✅ Regularly rotate credentials  

## Environment File Security

### File Structure:
```
.env                 # Real credentials (NEVER commit)
.env.example         # Template with placeholders (safe to commit)
.env.development     # Dev environment (NEVER commit)
.env.production      # Production environment (NEVER commit)
```

### Git Configuration:
The `.gitignore` file is configured to exclude:
- `.env` and `.env.*` files
- Certificate files (*.key, *.pem, *.crt)
- Any files containing "password", "secret", "credential"
- Database dumps and backups

## Script Security

### Database Scripts:
- Use `-p` flag without password to prompt securely
- Never store passwords in script variables
- Use arrays for command arguments to prevent shell expansion
- Redirect informational messages to stderr (>&2)

### Example Secure Usage:
```bash
# ✅ Secure - prompts for password
mysql -u $DB_USER -p -e "SELECT 1;"

# ❌ Insecure - password visible in process list
mysql -u $DB_USER -p$DB_PASSWORD -e "SELECT 1;"

# ✅ Secure - using array to prevent expansion
MYSQL_ARGS=(-u"$DB_USER" -p)
mysql "${MYSQL_ARGS[@]}" -e "SELECT 1;"
```

## Production Deployment

### Server Security:
- Run services as non-root users (www-data)
- Use systemd security features (ProtectSystem, NoNewPrivileges)
- Restrict file system access (ReadWritePaths)
- Use TLS/HTTPS for all web traffic
- Regularly update certificates

### Monitoring:
- Monitor log files for security events
- Set up alerts for failed login attempts
- Regularly audit file permissions
- Use fail2ban or similar for intrusion prevention

## Development Workflow

### Before Committing:
1. Run `git status` to check staged files
2. Verify no `.env` files are staged
3. Check for hardcoded credentials with: `grep -r "password\|secret" src/`
4. Use `git diff --cached` to review changes

### Environment Setup:
1. Copy `.env.example` to `.env`
2. Fill in actual credentials
3. Verify `.env` is in `.gitignore`
4. Never share your `.env` file

## Emergency Response

### If Credentials Are Accidentally Committed:
1. **IMMEDIATELY** change all affected passwords
2. Remove sensitive data from git history: `git filter-branch` or BFG
3. Force push to overwrite history: `git push --force-with-lease`
4. Notify team members to fetch latest changes
5. Audit access logs for unauthorized usage

### Contact Information:
- Security Issues: Contact system administrator immediately
- Password Reset: Use your organization's password management system

## Tools and Resources

### Recommended Tools:
- **Password Managers**: Bitwarden, 1Password, LastPass
- **Git Hooks**: Pre-commit hooks to scan for secrets
- **Scanning Tools**: git-secrets, truffleHog, GitLeaks
- **Environment Management**: direnv, docker-compose

### Additional Reading:
- OWASP Top 10 Security Risks
- NIST Cybersecurity Framework
- Your organization's security policies
