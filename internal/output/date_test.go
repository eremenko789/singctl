package output

import (
	"testing"
	"time"
)

func TestNormalizeLayout(t *testing.T) {
	if got := NormalizeLayout(""); got != DefaultDateLayout {
		t.Fatalf("empty: %q", got)
	}
	if got := NormalizeLayout("   "); got != DefaultDateLayout {
		t.Fatalf("blank: %q", got)
	}
	if got := NormalizeLayout("not-a-valid-layout-@@@"); got != DefaultDateLayout {
		t.Fatalf("invalid: %q", got)
	}
	if got := NormalizeLayout("02.01.2006"); got != "02.01.2006" {
		t.Fatalf("valid: %q", got)
	}
}

func TestFormatDate_TwoLayouts(t *testing.T) {
	tm := time.Date(2025, 11, 28, 0, 0, 0, 0, time.UTC)
	a := FormatDate(tm, "2006-01-02")
	b := FormatDate(tm, "02.01.2006")
	if a != "2025-11-28" {
		t.Fatalf("a=%q", a)
	}
	if b != "28.11.2025" {
		t.Fatalf("b=%q", b)
	}
	if a == b {
		t.Fatal("layouts must differ")
	}
	if got := FormatDate(tm, ""); got != "2025-11-28" {
		t.Fatalf("default empty layout: %q", got)
	}
}
