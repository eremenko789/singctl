package output

import "testing"

func TestColorEnabled(t *testing.T) {
	tests := []struct {
		name        string
		isTTY       bool
		noColorFlag bool
		noColorEnv  string
		configColor bool
		want        bool
	}{
		{"tty_ok", true, false, "", true, true},
		{"flag", true, true, "", true, false},
		{"env", true, false, "1", true, false},
		{"env_any", true, false, "yes", true, false},
		{"non_tty", false, false, "", true, false},
		{"config_off", true, false, "", false, false},
		{"empty_env_ok", true, false, "", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ColorEnabled(tt.isTTY, tt.noColorFlag, tt.noColorEnv, tt.configColor)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
