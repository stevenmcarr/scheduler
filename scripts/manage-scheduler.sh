#!/bin/bash
# WMU Course Scheduler Service Management Script

case "$1" in
    start)
        echo "Starting WMU Course Scheduler service..."
        sudo systemctl start scheduler.service
        sudo systemctl status scheduler.service --no-pager -l
        ;;
    stop)
        echo "Stopping WMU Course Scheduler service..."
        sudo systemctl stop scheduler.service
        echo "Service stopped."
        ;;
    restart)
        echo "Restarting WMU Course Scheduler service..."
        sudo systemctl restart scheduler.service
        sudo systemctl status scheduler.service --no-pager -l
        ;;
    status)
        sudo systemctl status scheduler.service --no-pager -l
        ;;
    logs)
        echo "=== Service Logs (last 50 lines) ==="
        sudo journalctl -u scheduler.service --no-pager -n 50
        echo ""
        echo "=== Application Logs (last 20 lines) ==="
        sudo tail -20 /var/log/scheduler/scheduler.log
        ;;
    enable)
        echo "Enabling WMU Course Scheduler to start on boot..."
        sudo systemctl enable scheduler.service
        ;;
    disable)
        echo "Disabling WMU Course Scheduler from starting on boot..."
        sudo systemctl disable scheduler.service
        ;;
    reload)
        echo "Reloading systemd and restarting service..."
        sudo systemctl daemon-reload
        sudo systemctl restart scheduler.service
        sudo systemctl status scheduler.service --no-pager -l
        ;;
    test)
        echo "Testing service connectivity..."
        echo -n "HTTP Status: "
        curl -s -o /dev/null -w "%{http_code}" http://localhost:4100/scheduler/
        echo ""
        echo "Making test request..."
        response=$(curl -s -H "User-Agent: ManagementScript/1.0" http://localhost:4100/scheduler/test)
        if [[ $? -eq 0 ]]; then
            echo "✅ Service is responding successfully"
            echo "Latest HTTP log entry:"
            sudo tail -1 /var/log/scheduler/scheduler.log | grep "HTTP"
        else
            echo "❌ Service is not responding"
        fi
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs|enable|disable|reload|test}"
        echo ""
        echo "Commands:"
        echo "  start    - Start the service"
        echo "  stop     - Stop the service"
        echo "  restart  - Restart the service"
        echo "  status   - Show service status"
        echo "  logs     - Show recent logs"
        echo "  enable   - Enable service to start on boot"
        echo "  disable  - Disable service from starting on boot"
        echo "  reload   - Reload systemd configuration and restart"
        echo "  test     - Test service connectivity"
        exit 1
        ;;
esac
