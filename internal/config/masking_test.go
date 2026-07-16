package config

import "testing"

func TestNormalizeStoredToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "bare token is preserved",
			input: "test-token-aaaa",
			want:  "test-token-aaaa",
		},
		{
			name:    "bearer prefix with whitespace is rejected",
			input:   "Bearer test-token-aaaa",
			wantErr: true,
		},
		{
			name:    "blank token is rejected",
			input:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeStoredToken(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("NormalizeStoredToken() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeStoredToken() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeStoredToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMaskToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "long token keeps edges only",
			input: "test-token-aaaa",
			want:  "test****aaaa",
		},
		{
			name:  "short token is fully hidden",
			input: "short",
			want:  "****",
		},
		{
			name:  "empty token remains empty",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MaskToken(tt.input); got != tt.want {
				t.Fatalf("MaskToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "adds bearer prefix",
			input: "test-token-aaaa",
			want:  "Bearer test-token-aaaa",
		},
		{
			name:  "keeps existing bearer prefix",
			input: "Bearer test-token-aaaa",
			want:  "Bearer Bearer test-token-aaaa",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := AuthorizationHeader(tt.input); got != tt.want {
				t.Fatalf("AuthorizationHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}
