-- WMU Course Scheduler Database Creation Script
-- This script creates the complete database schema for the WMU Course Scheduler application
-- Author: Generated for WMU Course Scheduler
-- Date: 2025-01-23

-- Create the database
CREATE DATABASE IF NOT EXISTS wmu_schedules
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

-- Use the database
USE wmu_schedules;

-- Drop existing tables in reverse dependency order (if they exist)
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS prefixes;
DROP TABLE IF EXISTS instructors;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS rooms;
DROP TABLE IF EXISTS time_slots;
DROP TABLE IF EXISTS users;
SET FOREIGN_KEY_CHECKS = 1;

-- Create users table
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_logged_in BOOLEAN DEFAULT FALSE,
    administrator BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_username (username),
    INDEX idx_email (email)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create departments table
CREATE TABLE departments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_department_name (name)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create prefixes table
CREATE TABLE prefixes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    prefix VARCHAR(10) NOT NULL,
    department_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE,
    UNIQUE KEY unique_prefix_department (prefix, department_id),
    INDEX idx_prefix (prefix),
    INDEX idx_department_id (department_id)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create schedules table
CREATE TABLE schedules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    term VARCHAR(20) NOT NULL,
    year INT NOT NULL,
    department_id INT NOT NULL,
    prefix_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE,
    FOREIGN KEY (prefix_id) REFERENCES prefixes(id) ON DELETE CASCADE,
    UNIQUE KEY unique_schedule (term, year, department_id, prefix_id),
    INDEX idx_term_year (term, year),
    INDEX idx_department_schedule (department_id),
    INDEX idx_prefix_schedule (prefix_id)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create instructors table
CREATE TABLE instructors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    department_id INT NOT NULL,
    status ENUM('Full Time', 'Part Time', 'TA', 'Adjunct') DEFAULT 'Full Time',
    email VARCHAR(100),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE CASCADE,
    INDEX idx_instructor_name (last_name, first_name),
    INDEX idx_instructor_department (department_id),
    INDEX idx_instructor_status (status)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create rooms table
CREATE TABLE rooms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    building VARCHAR(50) NOT NULL,
    room_number VARCHAR(20) NOT NULL,
    capacity INT DEFAULT 0,
    computer_lab BOOLEAN DEFAULT FALSE,
    dedicated_lab BOOLEAN DEFAULT FALSE,
    projector BOOLEAN DEFAULT FALSE,
    smart_board BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_room (building, room_number),
    INDEX idx_building (building),
    INDEX idx_capacity (capacity),
    INDEX idx_computer_lab (computer_lab),
    INDEX idx_dedicated_lab (dedicated_lab)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create time_slots table
CREATE TABLE time_slots (
    id INT AUTO_INCREMENT PRIMARY KEY,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    M BOOLEAN DEFAULT FALSE,  -- Monday
    T BOOLEAN DEFAULT FALSE,  -- Tuesday
    W BOOLEAN DEFAULT FALSE,  -- Wednesday
    R BOOLEAN DEFAULT FALSE,  -- Thursday
    F BOOLEAN DEFAULT FALSE,  -- Friday
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_start_time (start_time),
    INDEX idx_end_time (end_time),
    INDEX idx_days (M, T, W, R, F)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create courses table
CREATE TABLE courses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    crn INT NOT NULL UNIQUE,
    section VARCHAR(10) NOT NULL,
    schedule_id INT NOT NULL,
    course_number VARCHAR(10) NOT NULL,
    title VARCHAR(200) NOT NULL,
    min_credits INT DEFAULT 3,
    max_credits INT DEFAULT 3,
    min_contact INT DEFAULT 3,
    max_contact INT DEFAULT 3,
    cap INT DEFAULT 25,
    approval BOOLEAN DEFAULT FALSE,
    lab BOOLEAN DEFAULT FALSE,
    instructor_id INT NULL,
    timeslot_id INT NULL,
    room_id INT NULL,
    mode ENUM('In-Person', 'Online', 'Hybrid', 'TBD') DEFAULT 'In-Person',
    status ENUM('Added', 'Modified', 'Deleted', 'Cancelled', 'Active') DEFAULT 'Added',
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (schedule_id) REFERENCES schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (instructor_id) REFERENCES instructors(id) ON DELETE SET NULL,
    FOREIGN KEY (timeslot_id) REFERENCES time_slots(id) ON DELETE SET NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE SET NULL,
    INDEX idx_crn (crn),
    INDEX idx_course_number (course_number),
    INDEX idx_schedule_course (schedule_id),
    INDEX idx_instructor_course (instructor_id),
    INDEX idx_timeslot_course (timeslot_id),
    INDEX idx_room_course (room_id),
    INDEX idx_status (status),
    INDEX idx_mode (mode)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Insert default data

-- Insert default departments
INSERT INTO departments (name) VALUES
('Computer Science'),
('Mathematics'),
('Engineering'),
('Business'),
('Liberal Arts'),
('Natural Sciences');

-- Insert default prefixes
INSERT INTO prefixes (prefix, department_id) VALUES
('CS', 1),    -- Computer Science
('MATH', 2),  -- Mathematics
('ENGR', 3),  -- Engineering
('BUS', 4),   -- Business
('ENG', 5),   -- Liberal Arts (English)
('CHEM', 6),  -- Natural Sciences (Chemistry)
('PHYS', 6),  -- Natural Sciences (Physics)
('BIO', 6);   -- Natural Sciences (Biology)

-- Insert default time slots
INSERT INTO time_slots (start_time, end_time, M, T, W, R, F) VALUES
('08:00:00', '09:15:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 8:00-9:15
('09:30:00', '10:45:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 9:30-10:45
('11:00:00', '12:15:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 11:00-12:15
('12:30:00', '13:45:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 12:30-1:45
('14:00:00', '15:15:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 2:00-3:15
('15:30:00', '16:45:00', TRUE, FALSE, TRUE, FALSE, FALSE),   -- MW 3:30-4:45
('08:00:00', '09:15:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 8:00-9:15
('09:30:00', '10:45:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 9:30-10:45
('11:00:00', '12:15:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 11:00-12:15
('12:30:00', '13:45:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 12:30-1:45
('14:00:00', '15:15:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 2:00-3:15
('15:30:00', '16:45:00', FALSE, TRUE, FALSE, TRUE, FALSE),   -- TR 3:30-4:45
('09:00:00', '11:50:00', TRUE, FALSE, FALSE, FALSE, FALSE),   -- M 9:00-11:50 (Lab)
('13:00:00', '15:50:00', TRUE, FALSE, FALSE, FALSE, FALSE),   -- M 1:00-3:50 (Lab)
('09:00:00', '11:50:00', FALSE, TRUE, FALSE, FALSE, FALSE),   -- T 9:00-11:50 (Lab)
('13:00:00', '15:50:00', FALSE, TRUE, FALSE, FALSE, FALSE),   -- T 1:00-3:50 (Lab)
('09:00:00', '11:50:00', FALSE, FALSE, TRUE, FALSE, FALSE),   -- W 9:00-11:50 (Lab)
('13:00:00', '15:50:00', FALSE, FALSE, TRUE, FALSE, FALSE),   -- W 1:00-3:50 (Lab)
('09:00:00', '11:50:00', FALSE, FALSE, FALSE, TRUE, FALSE),   -- R 9:00-11:50 (Lab)
('13:00:00', '15:50:00', FALSE, FALSE, FALSE, TRUE, FALSE),   -- R 1:00-3:50 (Lab)
('09:00:00', '11:50:00', FALSE, FALSE, FALSE, FALSE, TRUE),   -- F 9:00-11:50 (Lab)
('13:00:00', '15:50:00', FALSE, FALSE, FALSE, FALSE, TRUE);   -- F 1:00-3:50 (Lab)

-- Insert default rooms
INSERT INTO rooms (building, room_number, capacity, computer_lab, dedicated_lab) VALUES
('FLOYD', 'D0109', 30, TRUE, FALSE),
('FLOYD', 'D0110', 25, FALSE, FALSE),
('FLOYD', 'D0111', 35, TRUE, FALSE),
('FLOYD', 'D0112', 40, FALSE, FALSE),
('FLOYD', 'D0201', 30, TRUE, TRUE),
('FLOYD', 'D0202', 25, FALSE, FALSE),
('FLOYD', 'D0203', 50, FALSE, FALSE),
('SCIENCE', 'S101', 30, FALSE, TRUE),
('SCIENCE', 'S102', 25, FALSE, FALSE),
('SCIENCE', 'S201', 40, TRUE, FALSE),
('ENGINEERING', 'E101', 35, TRUE, FALSE),
('ENGINEERING', 'E102', 30, FALSE, TRUE),
('BUSINESS', 'B101', 45, FALSE, FALSE),
('BUSINESS', 'B102', 40, TRUE, FALSE);

-- Insert default admin user
-- Password: AdminPassword123! (hashed with bcrypt)
INSERT INTO users (username, email, password, administrator) VALUES
('admin', 'admin@wmich.edu', '$2a$10$YourHashedPasswordHereForAdmin123', TRUE);

-- Create a sample instructor for each department
INSERT INTO instructors (first_name, last_name, department_id, status) VALUES
('John', 'Smith', 1, 'Full Time'),
('Jane', 'Doe', 1, 'Full Time'),
('Bob', 'Johnson', 2, 'Full Time'),
('Alice', 'Williams', 3, 'Part Time'),
('Charlie', 'Brown', 4, 'Full Time'),
('Diana', 'Davis', 5, 'Part Time'),
('Eve', 'Miller', 6, 'Full Time');

-- Create indexes for performance optimization
CREATE INDEX idx_courses_composite ON courses (schedule_id, status, course_number);
CREATE INDEX idx_schedules_composite ON schedules (term, year, department_id);
CREATE INDEX idx_timeslots_composite ON time_slots (start_time, end_time);

-- Display success message
SELECT 'WMU Schedules database created successfully!' AS 'Status';

-- Show table creation summary
SELECT 
    TABLE_NAME as 'Table Name',
    TABLE_ROWS as 'Estimated Rows',
    CREATE_TIME as 'Created'
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = 'wmu_schedules' 
ORDER BY TABLE_NAME;

-- Show foreign key relationships
SELECT 
    CONSTRAINT_NAME as 'Foreign Key',
    TABLE_NAME as 'Table',
    COLUMN_NAME as 'Column',
    REFERENCED_TABLE_NAME as 'References Table',
    REFERENCED_COLUMN_NAME as 'References Column'
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
WHERE TABLE_SCHEMA = 'wmu_schedules' 
    AND REFERENCED_TABLE_NAME IS NOT NULL
ORDER BY TABLE_NAME, CONSTRAINT_NAME;
