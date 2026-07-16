package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Kind classifies a user-facing API or client error for messaging and exit mapping.
type Kind string

// Kind constants (ТЗ §8.1 / F05 data-model).
const (
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindValidation   Kind = "validation"
	KindRateLimited  Kind = "rate_limited"
	KindServer       Kind = "server"
	KindOther        Kind = "other"
	KindConfig       Kind = "config"
	KindTransport    Kind = "transport"
	KindDate         Kind = "date"
)

// Stable catalog messages (ТЗ §8.1 / F05 data-model).
const (
	msgUnauthorized = "Error: invalid token. Run 'singctl config set-token'"
	msgForbidden    = "Error: insufficient token permissions"
	msgNotFound     = "Error: entity not found"
	msgValidation   = "Error: validation failed"
	msgRateLimited  = "Error: rate limited. Retry later"
	msgServer       = "Error: server error, retry later"
	msgTransport    = "Error: could not reach API"
)

// ClassifiedError is a user-facing error after HTTP/client taxonomy.
type ClassifiedError struct {
	Kind       Kind
	Message    string
	Cause      error
	EntityID   string
	StatusCode int
}

// Error returns the stable user-facing message.
func (e *ClassifiedError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap returns the underlying cause when present.
func (e *ClassifiedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

type classifyOpts struct {
	entityID string
}

// ClassifyOption configures Classify.
type ClassifyOption func(*classifyOpts)

// WithEntityID sets an optional entity id for 404 messages.
func WithEntityID(id string) ClassifyOption {
	return func(o *classifyOpts) {
		o.entityID = id
	}
}

// Classify maps errors to a ClassifiedError with Kind and catalog Message.
// nil → nil. Already-classified errors are returned as-is (via errors.As).
func Classify(err error, opts ...ClassifyOption) error {
	if err == nil {
		return nil
	}

	var existing *ClassifiedError
	if errors.As(err, &existing) {
		return existing
	}

	o := &classifyOpts{}
	for _, opt := range opts {
		opt(o)
	}

	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return classifyHTTP(httpErr, o.entityID)
	}

	var dateErr *DateError
	if errors.As(err, &dateErr) {
		msg := dateErr.Error()
		return &ClassifiedError{
			Kind:    KindDate,
			Message: msg,
			Cause:   err,
		}
	}

	if isMissingTokenErr(err) {
		return &ClassifiedError{
			Kind:    KindConfig,
			Message: err.Error(),
			Cause:   err,
		}
	}

	return &ClassifiedError{
		Kind:    KindTransport,
		Message: msgTransport,
		Cause:   err,
	}
}

func classifyHTTP(httpErr *HTTPError, entityID string) *ClassifiedError {
	code := httpErr.StatusCode
	ce := &ClassifiedError{
		Cause:      httpErr,
		StatusCode: code,
		EntityID:   entityID,
	}

	switch {
	case code == 401:
		ce.Kind = KindUnauthorized
		ce.Message = msgUnauthorized
	case code == 403:
		ce.Kind = KindForbidden
		ce.Message = msgForbidden
	case code == 404:
		ce.Kind = KindNotFound
		if id := strings.TrimSpace(entityID); id != "" {
			ce.Message = msgNotFound + ": " + id
		} else {
			ce.Message = msgNotFound
		}
	case code == 422:
		ce.Kind = KindValidation
		ce.Message = extractValidationMessage(httpErr.Body)
	case code == 429:
		ce.Kind = KindRateLimited
		ce.Message = msgRateLimited
	case code >= 500 && code <= 599:
		ce.Kind = KindServer
		ce.Message = msgServer
	default:
		ce.Kind = KindOther
		ce.Message = fmt.Sprintf("Error: request failed (HTTP %d)", code)
	}
	return ce
}

func isMissingTokenErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.Contains(msg, "токен") {
		return false
	}
	return strings.Contains(msg, "не задан") || strings.Contains(msg, "set-token")
}

func extractValidationMessage(body []byte) string {
	const maxLen = 2048
	if len(body) == 0 {
		return msgValidation
	}
	if !utf8.Valid(body) {
		return msgValidation
	}
	s := strings.TrimSpace(string(body))
	if s == "" || !isMostlyPrintable(s) {
		return msgValidation
	}
	if len(s) > maxLen {
		s = s[:maxLen]
	}

	var obj map[string]json.RawMessage
	if json.Unmarshal([]byte(s), &obj) == nil {
		for _, key := range []string{"message", "error", "detail"} {
			raw, ok := obj[key]
			if !ok {
				continue
			}
			var str string
			if json.Unmarshal(raw, &str) == nil {
				str = strings.TrimSpace(str)
				if str != "" {
					return str
				}
			}
		}
	}
	return s
}

func isMostlyPrintable(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			continue
		}
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// HTTPError is a typed non-2xx HTTP response error for callers (e.g. F05 taxonomy).
type HTTPError struct {
	StatusCode int
	Body       []byte
}

// Error implements the error interface without embedding secrets from the body.
func (e *HTTPError) Error() string {
	if e == nil {
		return "ошибка HTTP"
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// EnsureSuccess returns nil for 2xx status codes; otherwise an *HTTPError.
func EnsureSuccess(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode <= 299 {
		return nil
	}
	return &HTTPError{
		StatusCode: statusCode,
		Body:       body,
	}
}
