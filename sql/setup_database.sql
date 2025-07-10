-- Create the wmu_schedules database
CREATE DATABASE IF NOT EXISTS wmu_schedules;

-- Create the wmu_cs user with password
CREATE USER IF NOT EXISTS 'wmu_cs'@'localhost' IDENTIFIED BY '1h0ck3y$';

-- Grant all privileges on the wmu_schedules database to wmu_cs user
GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'localhost';

-- Also create user for any host (optional, for remote access)
CREATE USER IF NOT EXISTS 'wmu_cs'@'%' IDENTIFIED BY '1h0ck3y$';
GRANT ALL PRIVILEGES ON wmu_schedules.* TO 'wmu_cs'@'%';

-- Refresh the privileges
FLUSH PRIVILEGES;

-- Show databases to confirm creation
SHOW DATABASES;

-- Show users to confirm user creation
SELECT User, Host FROM mysql.user WHERE User = 'wmu_cs';

-- Use the database
USE wmu_schedules;

-- Show that we're in the correct database
SELECT DATABASE();
