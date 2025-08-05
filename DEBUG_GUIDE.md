# VS Code Debugging Guide for WMU Course Scheduler

This guide explains how to debug the WMU Course Scheduler application in VS Code.

## Launch Configurations

The project includes several launch configurations in `.vscode/launch.json`:

### 1. **Launch Go App in Chrome** (Default)
- Starts the application and opens Chrome automatically
- Uses production settings (HTTPS, port 8081)
- Best for full application testing

### 2. **Debug WMU Scheduler**
- Standard debugging configuration
- Uses your current `.env` file
- Runs in integrated terminal with detailed logging
- Good for general debugging

### 3. **Debug WMU Scheduler (Debug Env)**
- Uses the special `.env.debug` file
- Configured for easier debugging:
  - HTTP instead of HTTPS (no certificate warnings)
  - Port 8080 for easier access
  - Debug logging enabled
  - Debug mode for Gin framework

### 4. **Debug WMU Scheduler (No Browser)**
- Pure debugging without browser automation
- Best for backend debugging
- Console output in integrated terminal

## Debugging Features

### Working Directory
All launch configurations are set to run from the workspace root directory (`${workspaceFolder}`), ensuring:
- Proper access to `.env` files
- Correct template and static file loading
- Access to `certs/` directory for TLS certificates
- Proper relative path resolution

### Environment Variables
- **Production**: Uses `.env` file
- **Debug**: Uses `.env.debug` file (HTTP, different port, debug logging)
- **Custom**: Set `ENV_FILE` environment variable to specify a custom file

### Debugging Settings
The workspace includes optimized Go debugging settings:
- Language server enabled for better diagnostics
- Integrated terminal for console output
- Detailed logging and error reporting
- Proper Go toolchain configuration

## Available Tasks

Use `Ctrl+Shift+P` → "Tasks: Run Task" to access these tasks:

### Build Tasks
- **build**: Compile the application (`go build -o scheduler src/*.go`)
- **run**: Build and run the application
- **test**: Run Go tests

### Certificate Tasks
- **check-certs**: Check TLS certificate status
- **create-certs**: Create new self-signed certificates

### Browser Tasks
- **open-chrome**: Open application in Chrome (HTTPS on port 8081)

## Debug Environment Configuration

The `.env.debug` file is optimized for debugging:

```bash
# Simplified configuration for debugging
SERVER_PORT=8080          # Standard HTTP port
TLS_ENABLED=false         # No HTTPS complications
LOG_LEVEL=DEBUG           # Verbose logging
GIN_MODE=debug           # Gin debug mode
```

## How to Debug

### 1. **Start Debugging**
1. Open VS Code in the scheduler workspace
2. Go to Run and Debug view (`Ctrl+Shift+D`)
3. Select your preferred launch configuration
4. Press F5 or click "Start Debugging"

### 2. **Set Breakpoints**
- Click in the gutter next to line numbers to set breakpoints
- Breakpoints work in all `.go` files
- Use conditional breakpoints for specific scenarios

### 3. **Debug Controls**
- **F5**: Continue/Start
- **F10**: Step Over
- **F11**: Step Into
- **Shift+F11**: Step Out
- **Ctrl+Shift+F5**: Restart
- **Shift+F5**: Stop

### 4. **Debug Console**
- View variables and their values
- Execute Go expressions
- Inspect call stack
- View goroutines

## Common Debugging Scenarios

### Database Issues
1. Use "Debug WMU Scheduler (Debug Env)" configuration
2. Set breakpoints in `ConnectMySQL()` function
3. Check database connection parameters
4. Verify `.env.debug` database settings

### TLS Certificate Issues
1. Use "Debug WMU Scheduler (Debug Env)" (disables HTTPS)
2. Or use "check-certs" task to verify certificates
3. Set breakpoints in TLS detection logic in `main.go`

### Template/Static File Issues
1. Ensure debugging from workspace directory
2. Check relative paths in `routes.go`
3. Verify `templates/` and `images/` directories

### HTTP Request Issues
1. Set breakpoints in controller functions (`controllers.go`)
2. Use "Debug WMU Scheduler (Debug Env)" for HTTP access
3. Test with `curl http://localhost:8080/scheduler`

## Tips and Best Practices

### 1. **Use Debug Environment**
- Start with `.env.debug` configuration for simpler debugging
- Switch to production settings only when needed

### 2. **Monitor Logs**
- Watch the integrated terminal for application logs
- Look for AppLogger messages showing configuration choices

### 3. **Test Certificate Detection**
- Use the certificate management script: `./manage-certs.sh check`
- Test different certificate scenarios

### 4. **Database Debugging**
- Verify database connection first
- Check MySQL service is running
- Validate credentials in `.env` file

### 5. **Browser Testing**
- Debug environment uses HTTP on port 8080 (easier testing)
- Production uses HTTPS on port 8081
- Use `curl` for API testing without browser complications

## Troubleshooting

### "Failed to launch" errors
- Ensure Go extension is installed
- Verify Go toolchain is properly configured
- Check that `go` command is in PATH

### "Cannot find package" errors
- Run `go mod tidy` in workspace directory
- Ensure all dependencies are downloaded

### "Permission denied" for log files
- Normal in development environment
- Application falls back to stdout logging
- Set proper permissions for production: `sudo mkdir -p /var/log/scheduler && sudo chown $USER /var/log/scheduler`

### Database connection errors
- Verify MySQL is running
- Check credentials in `.env` or `.env.debug`
- Test connection manually: `mysql -u username -p database_name`

## File Structure for Debugging

```
/home/stevecarr/scheduler/          # Workspace root (debugging cwd)
├── .env                           # Production environment
├── .env.debug                     # Debug environment
├── src/                           # Go source files
│   ├── main.go                    # Main entry point
│   ├── controllers.go             # HTTP handlers
│   ├── routes.go                  # Route definitions
│   └── ...
├── certs/                         # TLS certificates
├── templates/                     # HTML templates
├── .vscode/                       # VS Code configuration
│   ├── launch.json               # Debug configurations
│   ├── tasks.json                # Build/run tasks
│   └── settings.json             # Workspace settings
└── manage-certs.sh               # Certificate management
```
