package output

import (
	"strings"
	"time"
)

// DefaultDateLayout is the fallback Go reference layout (YYYY-MM-DD).
const DefaultDateLayout = "2006-01-02"

// NormalizeLayout returns layout, or DefaultDateLayout when empty/invalid.
func NormalizeLayout(layout string) string {
	layout = strings.TrimSpace(layout)
	if layout == "" {
		return DefaultDateLayout
	}
	// Use a non-reference calendar date so Format(layout) != layout for valid layouts.
	probe := time.Date(2025, 11, 28, 15, 4, 5, 0, time.UTC)
	s := probe.Format(layout)
	if s == layout {
		// No reference components were substituted — treat as invalid.
		return DefaultDateLayout
	}
	if _, err := time.Parse(layout, s); err != nil {
		return DefaultDateLayout
	}
	return layout
}

// FormatDate formats t using layout (after NormalizeLayout).
func FormatDate(t time.Time, layout string) string {
	return t.Format(NormalizeLayout(layout))
}
