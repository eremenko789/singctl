package api_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/eremenko789/singctl/internal/api"
)

func TestParseDateAcceptsYYYYMMDD(t *testing.T) {
	t.Parallel()
	got, err := api.ParseDate("2025-11-28")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if got.Year() != 2025 || got.Month() != time.November || got.Day() != 28 {
		t.Fatalf("got %v", got)
	}
}

func TestTodayCalendarDateLocalYYYYMMDD(t *testing.T) {
	t.Parallel()
	got := api.TodayCalendarDate()
	if len(got) != 10 {
		t.Fatalf("len=%d got %q", len(got), got)
	}
	parsed, err := time.ParseInLocation("2006-01-02", got, time.Local)
	if err != nil {
		t.Fatalf("parse %q: %v", got, err)
	}
	now := time.Now().In(time.Local)
	if parsed.Year() != now.Year() || parsed.Month() != now.Month() || parsed.Day() != now.Day() {
		t.Fatalf("TodayCalendarDate=%q want local today %s", got, now.Format("2006-01-02"))
	}
}

func TestParseDateRejectsInvalid(t *testing.T) {
	t.Parallel()
	for _, in := range []string{"", "28.11.2025", "2025/11/28", "2025-13-01", "2025-02-30", "not-a-date"} {
		in := in
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			_, err := api.ParseDate(in)
			if err == nil {
				t.Fatal("expected error")
			}
			var de *api.DateError
			if !errors.As(err, &de) {
				t.Fatalf("want *DateError, got %T %v", err, err)
			}
			if !strings.Contains(err.Error(), "Expected: YYYY-MM-DD") {
				t.Fatalf("missing hint: %v", err)
			}

			classified := api.Classify(err)
			var ce *api.ClassifiedError
			if !errors.As(classified, &ce) {
				t.Fatalf("Classify: want *ClassifiedError, got %T", classified)
			}
			if ce.Kind != api.KindDate {
				t.Fatalf("Kind = %q, want date", ce.Kind)
			}
		})
	}
}
