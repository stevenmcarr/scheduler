# Course Conflict Detection

## Overview

The course conflict detection system identifies conflicts between courses with the same prefix that are scheduled at overlapping times. This feature helps ensure proper course scheduling by preventing conflicts within specific course number ranges.

## Course Conflict Rules

### 1. Course Number Ranges
Courses are grouped into the following ranges, and conflicts are detected within each range:

- **1000-1999**: Introductory/Freshman level courses
- **2000-2999**: Sophomore level courses  
- **3000-3999**: Junior level courses
- **5000-5999**: Graduate level courses
- **6000-6999**: Advanced graduate level courses

### 2. Conflict Detection Logic
Two courses with the same prefix will be flagged as conflicting if:

1. **Same Range**: Both course numbers fall within the same range (e.g., both in 2000-2999)
2. **Overlapping Time**: Their scheduled time slots overlap (same days and overlapping hours)
3. **No Exceptions**: They are not exempt due to the exceptions listed below

### 3. Exceptions (No Conflict)

Course conflicts are **NOT** reported in the following cases:

#### A. Crosslisted Courses
- Courses that are officially crosslisted can be scheduled at the same time
- This allows the same course to be offered under multiple prefixes simultaneously

#### B. Prerequisite Chain Courses  
- Courses that appear on the same prerequisite chain can be scheduled at the same time
- This includes both direct and indirect prerequisite relationships
- Example: If MATH 1050 → MATH 2150 → MATH 3150, then any two of these courses can be scheduled simultaneously

## Technical Implementation

### Database Integration
- Utilizes the existing `prerequisites` table with foreign key relationships to `prefixes`
- Leverages the `crosslistings` table for crosslisting exception checking
- Builds prerequisite chain graphs to detect indirect relationships

### Course Number Parsing
- Handles various course number formats (e.g., "2150", "2150H", "2150W")
- Extracts numeric portion for range comparison
- Supports honors (H), writing intensive (W), and other course variants

### Detection Algorithm
1. **Range Filtering**: Only compares courses within the same number range
2. **Time Overlap**: Checks for overlapping time slots using existing time conflict logic  
3. **Exception Checking**: 
   - Queries crosslisting table for crosslisted course pairs
   - Builds prerequisite chain graph and checks for relationships
4. **Conflict Reporting**: Generates structured conflict reports for display

## Usage

### Accessing Course Conflict Detection
1. Navigate to **Conflicts** → **Check Schedule Conflicts**
2. Select two schedules to compare
3. View the conflict report which includes:
   - Instructor conflicts
   - Room conflicts  
   - Crosslisting conflicts
   - **Course conflicts** (new feature)

### Interpreting Results
Course conflicts are displayed with:
- Course details (prefix, number, section, title, CRN)
- Time slot information
- Explanation of why the conflict was detected
- Clear visual distinction from other conflict types

## Example Scenarios

### Conflict Detected
- **CS 2150** (Data Structures) scheduled 10:00-11:00 MWF
- **CS 2200** (Computer Organization) scheduled 10:30-11:30 MWF
- **Result**: Conflict detected (both in 2000-2999 range with overlapping times)

### No Conflict - Different Ranges  
- **CS 1050** (Introduction to Programming) scheduled 10:00-11:00 MWF
- **CS 2150** (Data Structures) scheduled 10:00-11:00 MWF
- **Result**: No conflict (different ranges: 1000s vs 2000s)

### No Conflict - Prerequisite Chain
- **MATH 1050** (Algebra) scheduled 10:00-11:00 MWF  
- **MATH 1999** (Calculus Prep) scheduled 10:00-11:00 MWF
- **Result**: No conflict if MATH 1050 is a prerequisite for MATH 1999

### No Conflict - Crosslisted
- **CS 2150** scheduled 10:00-11:00 MWF
- **CET 2150** scheduled 10:00-11:00 MWF  
- **Result**: No conflict if these courses are crosslisted

## Benefits

1. **Academic Integrity**: Ensures students in the same academic level don't have scheduling conflicts
2. **Resource Planning**: Helps identify potential enrollment conflicts before they occur
3. **Automated Detection**: Reduces manual effort in conflict identification
4. **Flexible Exceptions**: Accommodates legitimate scheduling overlaps for related courses
5. **Comprehensive Reporting**: Provides detailed information for conflict resolution

## Future Enhancements

Potential improvements to the course conflict detection system:

1. **Custom Range Configuration**: Allow administrators to define custom course number ranges
2. **Department-Specific Rules**: Implement different conflict rules per department
3. **Priority-Based Conflicts**: Weight conflicts based on course importance or enrollment
4. **Semester Planning**: Extend conflict detection to multi-semester prerequisite planning
5. **Student Impact Analysis**: Show potential student enrollment conflicts
