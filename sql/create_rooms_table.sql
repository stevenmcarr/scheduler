-- Create rooms table in wmu_schedules database
USE wmu_schedules;

CREATE TABLE IF NOT EXISTS rooms (
    id INT AUTO_INCREMENT PRIMARY KEY,
    building VARCHAR(100) NOT NULL,
    room_number VARCHAR(20) NOT NULL,
    capacity INT NOT NULL,
    computer_lab BOOLEAN DEFAULT FALSE,
    dedicated_lab BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Add a unique constraint to prevent duplicate building/room_number combinations
    UNIQUE KEY unique_room (building, room_number)
);

-- Add indexes for better performance
CREATE INDEX idx_building ON rooms(building);
CREATE INDEX idx_room_number ON rooms(room_number);
CREATE INDEX idx_capacity ON rooms(capacity);
CREATE INDEX idx_computer_lab ON rooms(computer_lab);
CREATE INDEX idx_dedicated_lab ON rooms(dedicated_lab);
CREATE INDEX idx_labs ON rooms(computer_lab, dedicated_lab);

-- Show the table structure
DESCRIBE rooms;
