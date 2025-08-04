#!/bin/bash

# TLS Certificate Management Script for WMU Course Scheduler
# This script helps manage TLS certificates for the application

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="$SCRIPT_DIR/certs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=== WMU Course Scheduler TLS Certificate Manager ===${NC}"
    echo
}

check_certificates() {
    echo -e "${BLUE}Checking certificate status...${NC}"
    echo
    
    # Check environment-specified certificates
    if [ -n "$TLS_CERT_FILE" ] && [ -n "$TLS_KEY_FILE" ]; then
        if [ -f "$TLS_CERT_FILE" ] && [ -f "$TLS_KEY_FILE" ]; then
            echo -e "${GREEN}✓ Environment certificates found:${NC}"
            echo "  Cert: $TLS_CERT_FILE"
            echo "  Key:  $TLS_KEY_FILE"
            echo "  $(openssl x509 -in "$TLS_CERT_FILE" -noout -dates 2>/dev/null | head -1)"
            return 0
        else
            echo -e "${YELLOW}⚠ Environment certificates specified but not found:${NC}"
            echo "  Cert: $TLS_CERT_FILE"
            echo "  Key:  $TLS_KEY_FILE"
        fi
    fi
    
    # Check Let's Encrypt certificates
    LETSENCRYPT_CERT="/etc/letsencrypt/live/localhost/fullchain.pem"
    LETSENCRYPT_KEY="/etc/letsencrypt/live/localhost/privkey.pem"
    
    if [ -f "$LETSENCRYPT_CERT" ] && [ -f "$LETSENCRYPT_KEY" ]; then
        echo -e "${GREEN}✓ Let's Encrypt certificates found:${NC}"
        echo "  Cert: $LETSENCRYPT_CERT"
        echo "  Key:  $LETSENCRYPT_KEY"
        echo "  $(openssl x509 -in "$LETSENCRYPT_CERT" -noout -dates 2>/dev/null | head -1)"
        return 0
    else
        echo -e "${YELLOW}⚠ Let's Encrypt certificates not found${NC}"
    fi
    
    # Check self-signed certificates
    SELF_SIGNED_CERT="$CERT_DIR/server.crt"
    SELF_SIGNED_KEY="$CERT_DIR/server.key"
    
    if [ -f "$SELF_SIGNED_CERT" ] && [ -f "$SELF_SIGNED_KEY" ]; then
        echo -e "${GREEN}✓ Self-signed certificates found:${NC}"
        echo "  Cert: $SELF_SIGNED_CERT"
        echo "  Key:  $SELF_SIGNED_KEY"
        echo "  $(openssl x509 -in "$SELF_SIGNED_CERT" -noout -dates 2>/dev/null | head -1)"
        return 0
    else
        echo -e "${RED}✗ No self-signed certificates found${NC}"
        return 1
    fi
}

create_self_signed() {
    echo -e "${BLUE}Creating self-signed certificate...${NC}"
    
    # Create certs directory if it doesn't exist
    mkdir -p "$CERT_DIR"
    
    # Generate self-signed certificate
    openssl req -x509 -newkey rsa:4096 \
        -keyout "$CERT_DIR/server.key" \
        -out "$CERT_DIR/server.crt" \
        -days 365 -nodes \
        -subj "/C=US/ST=Michigan/L=Kalamazoo/O=Western Michigan University/OU=Computer Science/CN=localhost"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Self-signed certificate created successfully${NC}"
        echo "  Certificate: $CERT_DIR/server.crt"
        echo "  Private Key: $CERT_DIR/server.key"
        
        # Set appropriate permissions
        chmod 644 "$CERT_DIR/server.crt"
        chmod 600 "$CERT_DIR/server.key"
        
        echo -e "${YELLOW}Note: Self-signed certificates will show security warnings in browsers${NC}"
        echo -e "${YELLOW}      For production, use Let's Encrypt or a proper CA-signed certificate${NC}"
    else
        echo -e "${RED}✗ Failed to create self-signed certificate${NC}"
        return 1
    fi
}

view_certificate() {
    local cert_file="$1"
    if [ -f "$cert_file" ]; then
        echo -e "${BLUE}Certificate details for: $cert_file${NC}"
        openssl x509 -in "$cert_file" -text -noout | head -30
    else
        echo -e "${RED}Certificate file not found: $cert_file${NC}"
    fi
}

show_usage() {
    echo "Usage: $0 [command]"
    echo
    echo "Commands:"
    echo "  check      - Check certificate status"
    echo "  create     - Create self-signed certificate"
    echo "  view FILE  - View certificate details"
    echo "  help       - Show this help message"
    echo
    echo "If no command is specified, 'check' is run by default."
}

# Main script logic
print_header

case "${1:-check}" in
    "check")
        check_certificates
        ;;
    "create")
        create_self_signed
        ;;
    "view")
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please specify a certificate file to view${NC}"
            echo "Example: $0 view certs/server.crt"
            exit 1
        fi
        view_certificate "$2"
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo
        show_usage
        exit 1
        ;;
esac
