package api

import (
	"fmt"
	"time"
)

// DateError is returned when ParseDate rejects client input.
type DateError struct {
	Input   string
	Message string
}

// Error implements the error interface.
func (e *DateError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return "Error: invalid date. Expected: YYYY-MM-DD"
}

// ParseDate parses a calendar date in strict YYYY-MM-DD form (UTC midnight).
func ParseDate(s string) (time.Time, error) {
	const layout = "2006-01-02"
	const hint = "Expected: YYYY-MM-DD"
	if s == "" {
		return time.Time{}, &DateError{
			Input:   s,
			Message: fmt.Sprintf("Error: invalid date. %s", hint),
		}
	}
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, &DateError{
			Input:   s,
			Message: fmt.Sprintf("Error: invalid date. %s", hint),
		}
	}
	return t.UTC(), nil
}
