# Validation Package

## Overview
The validation package ensures that blood pressure and pulse readings are valid and consistent. It implements comprehensive validation rules based on medical guidelines and practical considerations.

## Validation Rules

### Range Checks
```go
const (
    MinSystolic  = 60
    MaxSystolic  = 250
    MinDiastolic = 40
    MaxDiastolic = 150
    MinPulse     = 40
    MaxPulse     = 200
)
```
- Enforces medically reasonable ranges for readings
- Prevents obvious data entry errors
- Constants make updates easy

### Consistency Checks
1. **Systolic vs Diastolic**
   - Systolic must be higher than diastolic
   - Medical requirement

2. **Reading Consistency**
   - Maximum 15 mmHg difference between readings
   - Ensures reliable measurements
   - Based on clinical guidelines

## Error Handling

### ValidationError Type
```go
type ValidationError struct {
    Field   string
    Message string
}
```
- Provides detailed error information
- Identifies specific problematic fields
- User-friendly error messages

### Multiple Errors
```go
type ValidationErrors []ValidationError
```
- Collects all validation errors
- Doesn't stop at first error
- Better user experience

## Go Concepts Demonstrated

1. **Custom Error Types**
   ```go
   func (e ValidationError) Error() string
   ```
   - Implements error interface
   - Provides custom error formatting
   - Type-safe error handling

2. **Variadic Functions**
   ```go
   func max(values ...int) int
   ```
   - Accepts variable number of arguments
   - Flexible helper functions
   - Common Go pattern

3. **Struct Slices**
   ```go
   readings := []struct {
       systolic  int
       diastolic int
       // ...
   }
   ```
   - Anonymous structs for temporary data
   - Clean iteration pattern

## Best Practices
1. Clear constant definitions
2. Comprehensive error messages
3. Multiple error collection
4. Separation of concerns
5. Helper functions for common operations
6. Panic only for programming errors

## Usage Example
```go
input := &models.ReadingInput{
    Systolic1:  120,
    Diastolic1: 80,
    Pulse1:     72,
    // ... other readings
}

if err := validation.ValidateReadings(input); err != nil {
    fmt.Println(err) // Prints all validation errors
    return
}
```

## Testing Guidelines
1. Test edge cases (minimum and maximum values)
2. Test invalid readings
3. Test reading consistency
4. Test error message formatting
5. Test multiple simultaneous errors
