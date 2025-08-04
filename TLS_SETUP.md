# TLS Certificate Setup Guide

The WMU Course Scheduler supports HTTPS/TLS encryption with automatic certificate detection and fallback mechanisms.

## Certificate Priority Order

The application checks for certificates in the following order:

1. **Environment-specified certificates** (via `TLS_CERT_FILE` and `TLS_KEY_FILE`)
2. **Let's Encrypt certificates** (in `/etc/letsencrypt/live/localhost/`)
3. **Self-signed certificates** (in `certs/` directory)

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# Enable TLS/HTTPS
TLS_ENABLED=true

# Optional: Specify custom certificate paths
TLS_CERT_FILE=/path/to/your/certificate.pem
TLS_KEY_FILE=/path/to/your/private-key.pem
```

If `TLS_CERT_FILE` and `TLS_KEY_FILE` are not specified or the files don't exist, the application will automatically check for Let's Encrypt certificates and fall back to self-signed certificates.

## Certificate Types

### 1. Let's Encrypt Certificates (Recommended for Production)

For production deployments, use Let's Encrypt for free, trusted certificates:

```bash
# Install certbot
sudo apt update
sudo apt install certbot

# Generate certificate (replace your-domain.com with your actual domain)
sudo certbot certonly --standalone -d your-domain.com

# Certificates will be stored in:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

Update your `.env` file:
```bash
TLS_CERT_FILE=/etc/letsencrypt/live/your-domain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/your-domain.com/privkey.pem
```

### 2. Self-Signed Certificates (Development/Testing)

For development or internal use, self-signed certificates are automatically created:

```bash
# Generate self-signed certificate (done automatically)
./manage-certs.sh create
```

**Note**: Self-signed certificates will show security warnings in browsers. They provide encryption but no identity verification.

## Certificate Management Script

Use the `manage-certs.sh` script to manage certificates:

```bash
# Check certificate status
./manage-certs.sh check

# Create new self-signed certificate
./manage-certs.sh create

# View certificate details
./manage-certs.sh view certs/server.crt

# Show help
./manage-certs.sh help
```

## Automatic Certificate Detection

The application automatically detects and uses the best available certificate:

1. **Environment certificates**: If `TLS_CERT_FILE` and `TLS_KEY_FILE` are set and files exist
2. **Let's Encrypt**: Checks `/etc/letsencrypt/live/localhost/` (configurable in code)
3. **Self-signed**: Uses `certs/server.crt` and `certs/server.key`
4. **HTTP fallback**: If no certificates are found, falls back to HTTP with a warning

## Security Considerations

### Production Deployment
- Use Let's Encrypt or CA-signed certificates
- Ensure proper file permissions (600 for private keys)
- Set up automatic certificate renewal
- Use a reverse proxy (nginx/Apache) for additional security

### Development/Testing
- Self-signed certificates are fine for local development
- Add security exceptions in browsers as needed
- Don't use self-signed certificates in production

## File Permissions

Ensure proper permissions for certificate files:

```bash
# Certificate files (readable by all)
chmod 644 /path/to/certificate.pem

# Private key files (readable only by owner)
chmod 600 /path/to/private-key.pem
```

## Troubleshooting

### Common Issues

1. **"Permission denied" errors**
   - Check file permissions
   - Ensure the application user has read access to certificates

2. **"Certificate not found" warnings**
   - Run `./manage-certs.sh check` to verify certificate status
   - Check paths in `.env` file

3. **Browser security warnings**
   - Normal for self-signed certificates
   - Add security exceptions or use CA-signed certificates

4. **Let's Encrypt certificate renewal**
   - Set up automatic renewal with cron
   - Test with `certbot renew --dry-run`

### Logs

Check application logs for TLS-related messages:

```bash
tail -f /var/log/scheduler/scheduler.log | grep -E "(TLS|HTTPS|certificate)"
```

## Testing HTTPS

Once configured, test the HTTPS connection:

```bash
# Test with curl (self-signed certificate)
curl -k https://localhost:8080/scheduler

# Test with curl (trusted certificate)
curl https://your-domain.com:8080/scheduler

# Test certificate details
openssl s_client -connect localhost:8080 -servername localhost
```

## Certificate Expiration

Monitor certificate expiration:

```bash
# Check certificate expiration
openssl x509 -in /path/to/certificate.pem -noout -dates

# Let's Encrypt auto-renewal (recommended)
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```
