package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func TestExitCodeTable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"not found", api.Classify(api.EnsureSuccess(404, nil)), 3},
		{"config", api.Classify(errors.New("токен не задан; используйте 'singctl config set-token'")), 2},
		{"unauthorized", api.Classify(api.EnsureSuccess(401, nil)), 1},
		{"forbidden", api.Classify(api.EnsureSuccess(403, nil)), 1},
		{"validation", api.Classify(api.EnsureSuccess(422, nil)), 1},
		{"rate limited", api.Classify(api.EnsureSuccess(429, nil)), 1},
		{"server", api.Classify(api.EnsureSuccess(500, nil)), 1},
		{"raw HTTP 404", api.EnsureSuccess(404, nil), 3},
		{"transport", api.Classify(fmt.Errorf("connection refused")), 1},
		{"date", api.Classify(mustDateErr(t)), 1},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ExitCode(tc.err); got != tc.want {
				t.Fatalf("ExitCode = %d, want %d (err=%v)", got, tc.want, tc.err)
			}
		})
	}
}

func mustDateErr(t *testing.T) error {
	t.Helper()
	_, err := api.ParseDate("")
	if err == nil {
		t.Fatal("expected date error")
	}
	return err
}
