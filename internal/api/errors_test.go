package api_test

import (
	"errors"
	"strings"
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

func TestClassifyCatalog4014035xx(t *testing.T) {
	t.Parallel()
	cases := []struct {
		code int
		kind api.Kind
		msg  string
	}{
		{401, api.KindUnauthorized, "Error: invalid token. Run 'singctl config set-token'"},
		{403, api.KindForbidden, "Error: insufficient token permissions"},
		{500, api.KindServer, "Error: server error, retry later"},
		{502, api.KindServer, "Error: server error, retry later"},
		{503, api.KindServer, "Error: server error, retry later"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.msg[:20], func(t *testing.T) {
			t.Parallel()
			raw := api.EnsureSuccess(tc.code, []byte(`secret-should-not-leak`))
			err := api.Classify(raw)
			var ce *api.ClassifiedError
			if !errors.As(err, &ce) {
				t.Fatalf("want *ClassifiedError, got %T %v", err, err)
			}
			if ce.Kind != tc.kind {
				t.Fatalf("Kind = %q, want %q", ce.Kind, tc.kind)
			}
			if ce.Message != tc.msg {
				t.Fatalf("Message = %q, want %q", ce.Message, tc.msg)
			}
			if strings.Contains(ce.Message, "secret") || strings.Contains(ce.Error(), "secret") {
				t.Fatalf("token/body leaked in message: %q", ce.Message)
			}
			if strings.Contains(ce.Message, "test-token") {
				t.Fatalf("token leaked: %q", ce.Message)
			}
		})
	}
}

func TestClassify404WithAndWithoutEntityID(t *testing.T) {
	t.Parallel()
	raw := api.EnsureSuccess(404, nil)

	err := api.Classify(raw)
	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("want *ClassifiedError, got %T", err)
	}
	if ce.Kind != api.KindNotFound {
		t.Fatalf("Kind = %q", ce.Kind)
	}
	if ce.Message != "Error: entity not found" {
		t.Fatalf("Message = %q", ce.Message)
	}

	err = api.Classify(raw, api.WithEntityID("proj-42"))
	if !errors.As(err, &ce) {
		t.Fatalf("want *ClassifiedError, got %T", err)
	}
	if ce.Message != "Error: entity not found: proj-42" {
		t.Fatalf("Message = %q", ce.Message)
	}
}

func TestClassify422BodyExtract(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		body []byte
		want string
	}{
		{"json message", []byte(`{"message":"field x invalid"}`), "field x invalid"},
		{"json error", []byte(`{"error":"bad input"}`), "bad input"},
		{"json detail", []byte(`{"detail":"nope"}`), "nope"},
		{"plain text", []byte(`plain validation text`), "plain validation text"},
		{"empty", nil, "Error: validation failed"},
		{"binary", []byte{0xff, 0xfe, 0x00}, "Error: validation failed"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := api.Classify(api.EnsureSuccess(422, tc.body))
			var ce *api.ClassifiedError
			if !errors.As(err, &ce) {
				t.Fatalf("want *ClassifiedError, got %T", err)
			}
			if ce.Kind != api.KindValidation {
				t.Fatalf("Kind = %q", ce.Kind)
			}
			if ce.Message != tc.want {
				t.Fatalf("Message = %q, want %q", ce.Message, tc.want)
			}
		})
	}
}

func TestClassify429(t *testing.T) {
	t.Parallel()
	err := api.Classify(api.EnsureSuccess(429, nil))
	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("want *ClassifiedError, got %T", err)
	}
	if ce.Kind != api.KindRateLimited {
		t.Fatalf("Kind = %q", ce.Kind)
	}
	if ce.Message != "Error: rate limited. Retry later" {
		t.Fatalf("Message = %q", ce.Message)
	}
}

func TestClassifyNilAndUnwrap(t *testing.T) {
	t.Parallel()
	if api.Classify(nil) != nil {
		t.Fatal("Classify(nil) must be nil")
	}

	raw := api.EnsureSuccess(401, []byte(`body`))
	err := api.Classify(raw)

	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("errors.As ClassifiedError failed: %v", err)
	}
	var httpErr *api.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("errors.As HTTPError failed: %v", err)
	}
	if httpErr.StatusCode != 401 {
		t.Fatalf("StatusCode = %d", httpErr.StatusCode)
	}
}

func TestClassifyMissingTokenConfig(t *testing.T) {
	t.Parallel()
	err := api.Classify(errors.New("токен не задан; используйте 'singctl config set-token'"))
	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		t.Fatalf("want *ClassifiedError, got %T", err)
	}
	if ce.Kind != api.KindConfig {
		t.Fatalf("Kind = %q, want config", ce.Kind)
	}
	if !strings.Contains(ce.Message, "set-token") {
		t.Fatalf("Message missing set-token hint: %q", ce.Message)
	}

	_, sessErr := api.NewSession("https://example.invalid", "", "30s")
	classified := api.Classify(sessErr)
	if !errors.As(classified, &ce) {
		t.Fatalf("want *ClassifiedError from factory, got %T", classified)
	}
	if ce.Kind != api.KindConfig {
		t.Fatalf("factory empty token Kind = %q, want config", ce.Kind)
	}
}
