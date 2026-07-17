package api_test

import (
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func TestNormalizeProjectEmoji(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"1f49e", "1f49e", false},
		{"1F49E", "1f49e", false},
		{"💞", "1f49e", false},
		{"  💞  ", "1f49e", false},
		{"abcd", "abcd", false},
		{"ABCDEF12", "abcdef12", false},
		{"", "", true},
		{"   ", "", true},
		{"heart", "", true},
		{"ab", "", true},         // too short for hex, not non-ASCII
		{"abcdefghij", "", true}, // too long for hex
		{"💞💞", "", true},
		{"h", "", true},
		{"9", "", true},
	}
	for _, tc := range cases {
		got, err := api.NormalizeProjectEmoji(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("NormalizeProjectEmoji(%q): want error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeProjectEmoji(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("NormalizeProjectEmoji(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}
