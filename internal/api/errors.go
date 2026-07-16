package api

import "fmt"

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
