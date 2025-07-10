-- Create instructors table in wmu_schedules database
USE wmu_schedules;

CREATE TABLE IF NOT EXISTS instructors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    last_name VARCHAR(100) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    department_id INT NOT NULL,
    status ENUM('full time', 'part time', 'TA') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Foreign key constraint to link to departments table
    CONSTRAINT fk_instructor_department 
        FOREIGN KEY (department_id) 
        REFERENCES departments(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

-- Add indexes for better performance
CREATE INDEX idx_last_name ON instructors(last_name);
CREATE INDEX idx_first_name ON instructors(first_name);
CREATE INDEX idx_department_id_instructors ON instructors(department_id);
CREATE INDEX idx_status ON instructors(status);
CREATE INDEX idx_full_name ON instructors(last_name, first_name);

-- Show the table structure
DESCRIBE instructors;
