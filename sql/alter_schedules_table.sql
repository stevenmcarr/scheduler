-- Alter schedules table in wmu_schedules database
USE wmu_schedules;

-- First, let's see the current structure
DESCRIBE schedules;

-- Step 1: Add the new department_id column
ALTER TABLE schedules 
ADD COLUMN department_id INT;

-- Step 2: Add a new prefix_id column to link to prefixes table
ALTER TABLE schedules 
ADD COLUMN prefix_id INT;

-- Step 3: Update the new columns with data from existing records
-- First, populate prefix_id based on existing prefix values
UPDATE schedules s 
JOIN prefixes p ON s.prefix = p.prefix 
SET s.prefix_id = p.id;

-- Then, populate department_id based on the prefix relationship
UPDATE schedules s 
JOIN prefixes p ON s.prefix_id = p.id 
SET s.department_id = p.department_id;

-- Step 4: Drop the old prefix column
ALTER TABLE schedules DROP COLUMN prefix;

-- Step 5: Make the new columns NOT NULL and add foreign key constraints
ALTER TABLE schedules 
MODIFY COLUMN prefix_id INT NOT NULL,
MODIFY COLUMN department_id INT NOT NULL;

-- Step 6: Add foreign key constraints
ALTER TABLE schedules 
ADD CONSTRAINT fk_schedule_prefix 
    FOREIGN KEY (prefix_id) 
    REFERENCES prefixes(id) 
    ON DELETE CASCADE 
    ON UPDATE CASCADE;

ALTER TABLE schedules 
ADD CONSTRAINT fk_schedule_department 
    FOREIGN KEY (department_id) 
    REFERENCES departments(id) 
    ON DELETE CASCADE 
    ON UPDATE CASCADE;

-- Step 7: Drop the old unique constraint and create a new one
ALTER TABLE schedules DROP INDEX unique_schedule;
ALTER TABLE schedules 
ADD CONSTRAINT unique_schedule 
    UNIQUE KEY (term, year, prefix_id);

-- Step 8: Add indexes for better performance
CREATE INDEX idx_prefix_id ON schedules(prefix_id);
CREATE INDEX idx_department_id_schedules ON schedules(department_id);

-- Show the new table structure
DESCRIBE schedules;

-- Show the relationships
SELECT 
    s.id,
    s.term,
    s.year,
    p.prefix,
    d.name as department_name
FROM schedules s
JOIN prefixes p ON s.prefix_id = p.id
JOIN departments d ON s.department_id = d.id
ORDER BY s.year, s.term, d.name, p.prefix;
