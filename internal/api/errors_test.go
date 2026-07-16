package api_test

import (
	"errors"
	"testing"

	"github.com/eremenko789/singctl/internal/api"
)

func TestEnsureSuccess2xx(t *testing.T) {
	t.Parallel()
	for _, code := range []int{200, 201, 204, 299} {
		if err := api.EnsureSuccess(code, nil); err != nil {
			t.Fatalf("status %d: unexpected error %v", code, err)
		}
	}
}

func TestEnsureSuccessNon2xxHTTPError(t *testing.T) {
	t.Parallel()
	for _, code := range []int{401, 404, 500} {
		err := api.EnsureSuccess(code, []byte(`{"msg":"fail"}`))
		if err == nil {
			t.Fatalf("status %d: expected error", code)
		}
		var httpErr *api.HTTPError
		if !errors.As(err, &httpErr) {
			t.Fatalf("status %d: want *HTTPError, got %T %v", code, err, err)
		}
		if httpErr.StatusCode != code {
			t.Fatalf("StatusCode = %d, want %d", httpErr.StatusCode, code)
		}
		if string(httpErr.Body) != `{"msg":"fail"}` {
			t.Fatalf("Body = %q", httpErr.Body)
		}
	}
}
