# Utils Package

## Overview
The utils package contains utility functions and types used across the application. Currently, it focuses on blood pressure classification according to American Heart Association (AHA) guidelines.

## Blood Pressure Classification

### Categories
1. **Normal**
   - Systolic: < 120 mmHg
   - Diastolic: < 80 mmHg

2. **Elevated**
   - Systolic: 120-129 mmHg
   - Diastolic: < 80 mmHg

3. **Hypertension Stage 1**
   - Systolic: 130-139 mmHg
   - Diastolic: 80-89 mmHg

4. **Hypertension Stage 2**
   - Systolic: ≥ 140 mmHg
   - Diastolic: ≥ 90 mmHg

5. **Hypertensive Crisis**
   - Systolic: > 180 mmHg
   - Diastolic: > 120 mmHg

## Code Organization

### BPCategory Type
```go
type BPCategory struct {
    Name        string
    Description string
    Risk        string
}
```
- Represents a blood pressure classification
- Includes description and risk level
- Immutable predefined categories

### Key Functions

1. **ClassifyBP**
   ```go
   func ClassifyBP(systolic, diastolic int) BPCategory
   ```
   - Takes systolic and diastolic readings
   - Returns appropriate category
   - Order of checks matters (most severe first)

2. **GetRecommendation**
   ```go
   func GetRecommendation(category BPCategory) string
   ```
   - Provides health recommendations
   - Specific to each category
   - Includes emergency warnings when needed

## Go Concepts Demonstrated

1. **Package Variables**
   ```go
   var CategoryNormal = BPCategory{...}
   ```
   - Predefined, immutable categories
   - Available throughout the package
   - Clean, maintainable approach

2. **Switch Statements**
   ```go
   switch category.Name {
       case CategoryNormal.Name:
           return "..."
   }
   ```
   - Clean multiple condition handling
   - Type-safe comparison

3. **Value Types vs Pointers**
   - BPCategory is passed by value
   - Efficient for small, immutable structs
   - No need for pointers here

## Best Practices
1. Clear, descriptive names for types and functions
2. Immutable predefined categories
3. Order conditions from most to least severe
4. Provide helpful recommendations
5. Use switch statements for clarity
6. Include default cases for safety

## Usage Example
```go
sys, dia := 128, 85
category := utils.ClassifyBP(sys, dia)
recommendation := utils.GetRecommendation(category)
fmt.Printf("Category: %s\nRecommendation: %s\n", category.Name, recommendation)
```
