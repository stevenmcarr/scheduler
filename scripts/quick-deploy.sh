#!/bin/bash
# Quick deployment script for development iterations
# This is a simplified version for rapid deployment during development

set -e

PROJECT_ROOT="/home/stevecarr/scheduler"
DEPLOY_DIR="/var/www/html/scheduler"
SERVICE_NAME="scheduler.service"

echo "🚀 Quick Deploy Starting..."

# Stop service
echo "⏹️  Stopping service..."
sudo systemctl stop $SERVICE_NAME 2>/dev/null || echo "Service not running"

# Quick compile
echo "🔨 Compiling..."
cd "$PROJECT_ROOT"
go build -o "$DEPLOY_DIR/scheduler" ./src

# Update templates
echo "📄 Updating templates..."
sudo cp -r src/templates/* "$DEPLOY_DIR/templates/"

# Set permissions
echo "🔐 Setting permissions..."
sudo chown www-data:www-data "$DEPLOY_DIR/scheduler"
sudo chmod +x "$DEPLOY_DIR/scheduler"

# Ensure log directory exists and is writable
echo "📝 Ensuring log directory..."
sudo mkdir -p /var/log/scheduler
sudo chown www-data:www-data /var/log/scheduler
sudo chmod 755 /var/log/scheduler

# Start service
echo "▶️  Starting service..."
sudo systemctl start $SERVICE_NAME

# Check status
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "✅ Service started successfully!"
    echo "🌐 Test at: http://localhost:8080/scheduler/"
else
    echo "❌ Service failed to start"
    sudo journalctl -u $SERVICE_NAME --no-pager -n 5
fi
