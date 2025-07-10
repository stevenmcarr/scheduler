-- Create time_slots table in wmu_schedules database
USE wmu_schedules;

CREATE TABLE IF NOT EXISTS time_slots (
    id INT AUTO_INCREMENT PRIMARY KEY,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    M BOOLEAN DEFAULT FALSE,
    T BOOLEAN DEFAULT FALSE,
    W BOOLEAN DEFAULT FALSE,
    R BOOLEAN DEFAULT FALSE,
    F BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Add a unique constraint to prevent duplicate time/day combinations
    UNIQUE KEY unique_time_slot (start_time, end_time, M, T, W, R, F),
    
    -- Check constraint to ensure end_time is after start_time
    CONSTRAINT chk_time_order CHECK (end_time > start_time)
);

-- Add indexes for better performance
CREATE INDEX idx_start_time ON time_slots(start_time);
CREATE INDEX idx_end_time ON time_slots(end_time);
CREATE INDEX idx_monday ON time_slots(M);
CREATE INDEX idx_tuesday ON time_slots(T);
CREATE INDEX idx_wednesday ON time_slots(W);
CREATE INDEX idx_thursday ON time_slots(R);
CREATE INDEX idx_friday ON time_slots(F);
CREATE INDEX idx_time_range ON time_slots(start_time, end_time);

-- Show the table structure
DESCRIBE time_slots;
