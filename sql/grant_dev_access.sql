-- Create and grant full access to wmu_cs user for wmu_schedules_dev database
-- This script will create the user if it doesn't exist and grant all privileges

-- Create user if it doesn't exist (MySQL 8.0+ syntax)
-- Note: Replace 'your_password_here' with the actual password
CREATE USER IF NOT EXISTS 'wmu_cs'@'localhost' IDENTIFIED BY 'mypassword';
CREATE USER IF NOT EXISTS 'wmu_cs'@'127.0.0.1' IDENTIFIED BY 'mypassword';
CREATE USER IF NOT EXISTS 'wmu_cs'@'%' IDENTIFIED BY 'mypassword';

-- Grant all privileges on the wmu_schedules_dev database to wmu_cs user
GRANT ALL PRIVILEGES ON wmu_schedules_dev.* TO 'wmu_cs'@'localhost';
GRANT ALL PRIVILEGES ON wmu_schedules_dev.* TO 'wmu_cs'@'127.0.0.1';
GRANT ALL PRIVILEGES ON wmu_schedules_dev.* TO 'wmu_cs'@'%';

-- Also grant privileges on the production database if needed
-- GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'localhost';
-- GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'127.0.0.1';
-- GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'%';

-- Reload the grant tables to ensure changes take effect
FLUSH PRIVILEGES;

-- Display current privileges for verification
SHOW GRANTS FOR 'wmu_cs'@'localhost';
SHOW GRANTS FOR 'wmu_cs'@'127.0.0.1';
SHOW GRANTS FOR 'wmu_cs'@'%';
