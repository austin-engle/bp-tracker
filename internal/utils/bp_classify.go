// File: internal/utils/bp_classify.go

package utils

import "fmt"

// BPCategory represents blood pressure classification categories
type BPCategory struct {
    Name        string
    Description string
    Risk        string
}

var (
    CategoryNormal = BPCategory{
        Name:        "Normal",
        Description: "Blood pressure in normal range",
        Risk:        "low",
    }
    CategoryElevated = BPCategory{
        Name:        "Elevated",
        Description: "Blood pressure is slightly high",
        Risk:        "moderate",
    }
    CategoryStage1 = BPCategory{
        Name:        "Hypertension Stage 1",
        Description: "Blood pressure is high",
        Risk:        "high",
    }
    CategoryStage2 = BPCategory{
        Name:        "Hypertension Stage 2",
        Description: "Blood pressure is very high",
        Risk:        "very high",
    }
    CategoryCrisis = BPCategory{
        Name:        "Hypertensive Crisis",
        Description: "Seek emergency medical attention",
        Risk:        "severe",
    }
)

// ClassifyBP determines the blood pressure category based on systolic and diastolic readings
func ClassifyBP(systolic, diastolic int) BPCategory {
    // Check for hypertensive crisis first (most severe)
    if systolic > 180 || diastolic > 120 {
        return CategoryCrisis
    }

    // Check for Stage 2 Hypertension
    if systolic >= 140 || diastolic >= 90 {
        return CategoryStage2
    }

    // Check for Stage 1 Hypertension
    if systolic >= 130 || diastolic >= 80 {
        return CategoryStage1
    }

    // Check for Elevated
    if systolic >= 120 && diastolic < 80 {
        return CategoryElevated
    }

    // If none of the above, it's normal
    return CategoryNormal
}

// GetRecommendation provides health recommendations based on blood pressure category
func GetRecommendation(category BPCategory) string {
    switch category.Name {
    case CategoryNormal.Name:
        return "Maintain a healthy lifestyle with regular exercise and balanced diet."

    case CategoryElevated.Name:
        return "Consider lifestyle changes including reduced sodium intake and regular exercise. Monitor BP regularly."

    case CategoryStage1.Name:
        return "Consult your healthcare provider. Lifestyle changes and possibly medication may be needed."

    case CategoryStage2.Name:
        return "Consult your healthcare provider promptly. Medication is likely needed along with lifestyle changes."

    case CategoryCrisis.Name:
        return "SEEK EMERGENCY MEDICAL ATTENTION IMMEDIATELY!"

    default:
        return fmt.Sprintf("Unknown category: %s. Please consult your healthcare provider.", category.Name)
    }
}
