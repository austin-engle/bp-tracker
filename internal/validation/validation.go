// File: internal/validation/validation.go

package validation

import (
	"bp-tracker/internal/models"
	"fmt"
)

// ValidationError represents an error in input validation
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a slice of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
    if len(e) == 0 {
        return ""
    }

    result := "Validation errors:\n"
    for _, err := range e {
        result += fmt.Sprintf("- %s\n", err.Error())
    }
    return result
}

// Ranges for valid blood pressure and pulse readings
const (
    MinSystolic  = 60
    MaxSystolic  = 250
    MinDiastolic = 40
    MaxDiastolic = 150
    MinPulse     = 40
    MaxPulse     = 200

    // Maximum allowed difference between readings
    MaxReadingDiff = 15
)

// ValidateReading checks if a single set of readings is within acceptable ranges
func validateSingleReading(systolic, diastolic, pulse int, readingNum int) ValidationErrors {
    var errors ValidationErrors

    // Check systolic range
    if systolic < MinSystolic || systolic > MaxSystolic {
        errors = append(errors, ValidationError{
            Field:   fmt.Sprintf("Systolic Reading %d", readingNum),
            Message: fmt.Sprintf("must be between %d and %d", MinSystolic, MaxSystolic),
        })
    }

    // Check diastolic range
    if diastolic < MinDiastolic || diastolic > MaxDiastolic {
        errors = append(errors, ValidationError{
            Field:   fmt.Sprintf("Diastolic Reading %d", readingNum),
            Message: fmt.Sprintf("must be between %d and %d", MinDiastolic, MaxDiastolic),
        })
    }

    // Check pulse range
    if pulse < MinPulse || pulse > MaxPulse {
        errors = append(errors, ValidationError{
            Field:   fmt.Sprintf("Pulse Reading %d", readingNum),
            Message: fmt.Sprintf("must be between %d and %d", MinPulse, MaxPulse),
        })
    }

    // Check systolic is higher than diastolic
    if systolic <= diastolic {
        errors = append(errors, ValidationError{
            Field:   fmt.Sprintf("Reading %d", readingNum),
            Message: "systolic pressure must be higher than diastolic pressure",
        })
    }

    return errors
}

// ValidateReadings validates all three readings
func ValidateReadings(input *models.ReadingInput) error {
    var allErrors ValidationErrors

    // Validate each reading individually
    readings := []struct {
        systolic  int
        diastolic int
        pulse     int
        num       int
    }{
        {input.Systolic1, input.Diastolic1, input.Pulse1, 1},
        {input.Systolic2, input.Diastolic2, input.Pulse2, 2},
        {input.Systolic3, input.Diastolic3, input.Pulse3, 3},
    }

    for _, r := range readings {
        if errs := validateSingleReading(r.systolic, r.diastolic, r.pulse, r.num); len(errs) > 0 {
            allErrors = append(allErrors, errs...)
        }
    }

    // Check consistency between readings
    if len(allErrors) == 0 {
        // Check systolic consistency
        maxSys := max(input.Systolic1, input.Systolic2, input.Systolic3)
        minSys := min(input.Systolic1, input.Systolic2, input.Systolic3)
        if maxSys-minSys > MaxReadingDiff {
            allErrors = append(allErrors, ValidationError{
                Field:   "Systolic Readings",
                Message: fmt.Sprintf("difference between readings cannot exceed %d mmHg", MaxReadingDiff),
            })
        }

        // Check diastolic consistency
        maxDia := max(input.Diastolic1, input.Diastolic2, input.Diastolic3)
        minDia := min(input.Diastolic1, input.Diastolic2, input.Diastolic3)
        if maxDia-minDia > MaxReadingDiff {
            allErrors = append(allErrors, ValidationError{
                Field:   "Diastolic Readings",
                Message: fmt.Sprintf("difference between readings cannot exceed %d mmHg", MaxReadingDiff),
            })
        }
    }

    if len(allErrors) > 0 {
        return allErrors
    }
    return nil
}

// Helper functions for finding min/max values
func max(values ...int) int {
    if len(values) == 0 {
        panic("no values provided to max")
    }
    m := values[0]
    for _, v := range values[1:] {
        if v > m {
            m = v
        }
    }
    return m
}

func min(values ...int) int {
    if len(values) == 0 {
        panic("no values provided to min")
    }
    m := values[0]
    for _, v := range values[1:] {
        if v < m {
            m = v
        }
    }
    return m
}
