-- Create prefixes table in wmu_schedules database
USE wmu_schedules;

CREATE TABLE IF NOT EXISTS prefixes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    prefix VARCHAR(10) NOT NULL UNIQUE,
    department_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Foreign key constraint to link to departments table
    CONSTRAINT fk_prefix_department 
        FOREIGN KEY (department_id) 
        REFERENCES departments(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

-- Add indexes for better performance
CREATE INDEX idx_prefix ON prefixes(prefix);
CREATE INDEX idx_department_id ON prefixes(department_id);

-- Show the table structure
DESCRIBE prefixes;
