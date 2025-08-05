#!/bin/bash
# Quick reference for users table structure

echo "=== WMU Scheduler Users Table Structure ==="
echo
echo "Actual columns in users table:"
echo "  id               INT AUTO_INCREMENT PRIMARY KEY"
echo "  username         VARCHAR(50) NOT NULL UNIQUE"
echo "  email            VARCHAR(100) NOT NULL UNIQUE"  
echo "  password         VARCHAR(255) NOT NULL"
echo "  is_logged_in     BOOLEAN DEFAULT FALSE"
echo "  administrator    BOOLEAN DEFAULT FALSE"
echo "  created_at       TIMESTAMP"
echo "  updated_at       TIMESTAMP"
echo
echo "Key differences from generic assumptions:"
echo "  ❌ No 'role' column → Use 'administrator' (BOOLEAN)"
echo "  ❌ No 'password_hash' → Use 'password'"
echo "  ✅ 'administrator' = 1 for admin, 0 for regular user"
echo
echo "Correct SQL patterns:"
echo "  WHERE administrator = 1     # For admins"
echo "  WHERE administrator != 1    # For non-admins"
echo "  UPDATE users SET password = '...' # Not password_hash"
