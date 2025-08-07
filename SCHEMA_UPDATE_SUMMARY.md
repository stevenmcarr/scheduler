# Prerequisites Table Schema Update Summary

## Changes Made

The prerequisites table has been updated to use foreign key references to the prefixes table instead of storing prefix strings directly.

### New Table Structure

```sql
CREATE TABLE prerequisites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    pred_prefix_id INT NOT NULL,           -- Foreign key to prefixes.id
    pred_course_num VARCHAR(10) NOT NULL,  -- Course number (e.g., "101")
    succ_prefix_id INT NOT NULL,           -- Foreign key to prefixes.id  
    succ_course_num VARCHAR(10) NOT NULL,  -- Course number (e.g., "201")
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_pred_prefix FOREIGN KEY (pred_prefix_id) REFERENCES prefixes(id),
    CONSTRAINT fk_succ_prefix FOREIGN KEY (succ_prefix_id) REFERENCES prefixes(id),
    
    -- Indexes for performance
    INDEX idx_predecessor (pred_prefix_id, pred_course_num),
    INDEX idx_successor (succ_prefix_id, succ_course_num),
    
    -- Unique constraint to prevent duplicate prerequisites
    UNIQUE KEY unique_prerequisite (pred_prefix_id, pred_course_num, succ_prefix_id, succ_course_num)
);
```

### Column Mapping

| Old Column Name      | New Column Name   | Type     | Description |
|---------------------|------------------|----------|-------------|
| predecessor_prefix   | pred_prefix_id   | INT      | Foreign key to prefixes table |
| predecessor_number   | pred_course_num  | VARCHAR  | Course number |
| successor_prefix     | succ_prefix_id   | INT      | Foreign key to prefixes table |
| successor_number     | succ_course_num  | VARCHAR  | Course number |

### Database Changes

#### 1. Updated Prerequisites Table Schema
- Changed from storing prefix strings to storing prefix IDs
- Added foreign key constraints to ensure data integrity
- Maintains indexes for performance
- Preserves unique constraint functionality

#### 2. Enhanced Prerequisite Struct
```go
type Prerequisite struct {
    ID                int
    PredPrefixID      int    // Database fields
    PredCourseNum     string
    SuccPrefixID      int
    SuccCourseNum     string
    // Display fields (populated from JOINs)
    PredecessorPrefix string
    PredecessorNumber string
    SuccessorPrefix   string
    SuccessorNumber   string
}
```

#### 3. Updated Database Functions
- **GetAllPrerequisites()**: Now uses JOINs to get both IDs and display strings
- **GetPrerequisitesByFilter()**: Updated with JOIN queries for filtering
- **AddPrerequisite()**: Converts prefix strings to IDs before insertion
- **UpdatePrerequisite()**: Converts prefix strings to IDs before update
- **GetPrefixID()**: Helper function to get prefix ID from string (already existed)

### Sample Data

The system includes sample prerequisites using the new schema:
```sql
-- Sample data with prefix IDs (assuming MATH=1, ENG=2, CS=3, PHYS=4, CHEM=5)
INSERT INTO prerequisites (pred_prefix_id, pred_course_num, succ_prefix_id, succ_course_num) VALUES
(1, '101', 1, '102'),  -- MATH 101 -> MATH 102
(1, '102', 1, '201'),  -- MATH 102 -> MATH 201
(3, '101', 3, '201'),  -- CS 101 -> CS 201
(1, '101', 3, '201');  -- MATH 101 -> CS 201 (cross-department)
```

### Benefits of New Schema

1. **Data Integrity**: Foreign key constraints ensure only valid prefixes are used
2. **Normalization**: Eliminates prefix string duplication
3. **Performance**: Integer JOINs are faster than string comparisons
4. **Consistency**: Ensures prefix spelling consistency across the system
5. **Referential Integrity**: Prevents orphaned prerequisite records

### Migration Notes

- **Backward Compatibility**: Web interface continues to work with prefix strings
- **Automatic Conversion**: Database functions handle prefix string ↔ ID conversion
- **Error Handling**: Graceful handling of invalid prefix strings
- **Fallback Support**: GetUniquePrefixes() provides fallback if prefixes table doesn't exist

### Files Modified

1. **create_prerequisites_table.sql**: Updated table schema and sample data
2. **src/db.go**: Updated Prerequisite struct and all database functions
3. **setup_prerequisites.sh**: Enhanced setup script with better messaging

### Testing the Changes

1. Run the setup script: `./setup_prerequisites.sh`
2. Build the application: `go build -o bin/course_scheduler src/*.go`
3. Access the interface: `http://localhost:8080/scheduler/prerequisites`

The system maintains full functionality while providing better data integrity and performance through the normalized schema.
