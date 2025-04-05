// File: internal/models/reading.go

package models

import (
    "time"
)

// Reading represents a single blood pressure reading session
type Reading struct {
    ID         int64     `json:"id"`
    Timestamp  time.Time `json:"timestamp"`
    Systolic   int       `json:"systolic"`
    Diastolic  int       `json:"diastolic"`
    Pulse      int       `json:"pulse"`
    Classification string `json:"classification"`
}

// ReadingInput represents the user input for three consecutive readings
type ReadingInput struct {
    // Optional timestamp
    Timestamp  string `json:"timestamp,omitempty"`

    // First Reading
    Systolic1  int `json:"systolic1"`
    Diastolic1 int `json:"diastolic1"`
    Pulse1     int `json:"pulse1"`

    // Second Reading
    Systolic2  int `json:"systolic2"`
    Diastolic2 int `json:"diastolic2"`
    Pulse2     int `json:"pulse2"`

    // Third Reading
    Systolic3  int `json:"systolic3"`
    Diastolic3 int `json:"diastolic3"`
    Pulse3     int `json:"pulse3"`
}

// Average calculates the average of three readings
func (ri *ReadingInput) Average() *Reading {
    r := &Reading{
        Systolic:  (ri.Systolic1 + ri.Systolic2 + ri.Systolic3) / 3,
        Diastolic: (ri.Diastolic1 + ri.Diastolic2 + ri.Diastolic3) / 3,
        Pulse:     (ri.Pulse1 + ri.Pulse2 + ri.Pulse3) / 3,
    }

    // Parse timestamp if provided, otherwise use current time
    if ri.Timestamp != "" {
        if t, err := time.Parse("2006-01-02 15:04:05", ri.Timestamp); err == nil {
            r.Timestamp = t
        } else {
            r.Timestamp = time.Now()
        }
    } else {
        r.Timestamp = time.Now()
    }

    return r
}

// GetTimestampInMST returns the current time in Mountain Standard Time
func GetTimestampInMST() time.Time {
    loc, _ := time.LoadLocation("America/Denver")
    return time.Now().In(loc)
}

// Stats represents the statistical data for blood pressure readings
type Stats struct {
    LastReading    *Reading `json:"last_reading"`
    SevenDayAvg    *Reading `json:"seven_day_avg"`
    SevenDayCount  int      `json:"seven_day_count"`
    ThirtyDayAvg   *Reading `json:"thirty_day_avg"`
    ThirtyDayCount int      `json:"thirty_day_count"`
    AllTimeAvg     *Reading `json:"all_time_avg"`
    AllTimeCount   int      `json:"all_time_count"`
}
