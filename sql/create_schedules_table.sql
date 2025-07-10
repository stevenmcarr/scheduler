-- Create schedules table in wmu_schedules database
USE wmu_schedules;

CREATE TABLE IF NOT EXISTS schedules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    term ENUM('Fall', 'Spring', 'Summer I', 'Summer II') NOT NULL,
    year YEAR NOT NULL,
    prefix VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Add a unique constraint to prevent duplicate term/year/prefix combinations
    UNIQUE KEY unique_schedule (term, year, prefix)
);

-- Add indexes for better performance
CREATE INDEX idx_term ON schedules(term);
CREATE INDEX idx_year ON schedules(year);
CREATE INDEX idx_prefix ON schedules(prefix);
CREATE INDEX idx_term_year ON schedules(term, year);

-- Show the table structure
DESCRIBE schedules;
