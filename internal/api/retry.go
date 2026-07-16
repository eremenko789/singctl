package api

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	maxRetryAttempts = 3
	retryAfterCap    = 30 * time.Second
)

// SleepFunc waits for d or until ctx is cancelled.
type SleepFunc func(ctx context.Context, d time.Duration) error

var (
	sleepMu   sync.RWMutex
	sleepFunc SleepFunc = defaultSleep
)

// SetSleepForTest replaces the package default sleeper; restore with the returned func.
// Prefer WithSleep on NewSession when possible; this exists for CLI integration tests.
func SetSleepForTest(fn SleepFunc) (restore func()) {
	sleepMu.Lock()
	prev := sleepFunc
	if fn == nil {
		sleepFunc = defaultSleep
	} else {
		sleepFunc = fn
	}
	sleepMu.Unlock()
	return func() {
		sleepMu.Lock()
		sleepFunc = prev
		sleepMu.Unlock()
	}
}

func currentSleep() SleepFunc {
	sleepMu.RLock()
	defer sleepMu.RUnlock()
	return sleepFunc
}

func defaultSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

type retryTransport struct {
	base  http.RoundTripper
	sleep SleepFunc
}

// RoundTrip retries only HTTP 429 up to maxRetryAttempts.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	sleep := t.sleep
	if sleep == nil {
		sleep = defaultSleep
	}

	var resp *http.Response
	var err error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		r := req
		if attempt > 1 {
			r = cloneRequest(req)
		}
		resp, err = base.RoundTrip(r)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt == maxRetryAttempts {
			return resp, nil
		}

		delay := retryDelay(resp, attempt)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if err := sleep(req.Context(), delay); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func retryDelay(resp *http.Response, completedAttempt int) time.Duration {
	if resp != nil {
		if d, ok := parseRetryAfter(resp.Header.Get("Retry-After")); ok {
			if d > retryAfterCap {
				return retryAfterCap
			}
			if d < 0 {
				return 0
			}
			return d
		}
	}
	// Before 2nd attempt: 1s; before 3rd: 2s.
	return time.Duration(1<<uint(completedAttempt-1)) * time.Second
}

func parseRetryAfter(v string) (time.Duration, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0, false
		}
		return time.Duration(secs) * time.Second, true
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0, false
		}
		return d, true
	}
	return 0, false
}

func cloneRequest(req *http.Request) *http.Request {
	r := req.Clone(req.Context())
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err == nil {
			r.Body = body
		}
	}
	return r
}
