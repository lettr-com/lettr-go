package lettr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// tooManyRequests answers every request with a 429 carrying errorCode (omitted
// when empty) and, when retryAfter is non-empty, a Retry-After header.
func tooManyRequests(errorCode, retryAfter string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if retryAfter != "" {
			w.Header().Set("Retry-After", retryAfter)
		}
		w.WriteHeader(http.StatusTooManyRequests)
		body := map[string]any{"message": "Too Many Requests"}
		if errorCode != "" {
			body["error_code"] = errorCode
		}
		_ = json.NewEncoder(w).Encode(body)
	}
}

func sendAgainst(t *testing.T, handler http.HandlerFunc) error {
	t.Helper()
	client, server := newTestClient(t, handler)
	defer server.Close()
	_, err := client.Emails.Send(context.Background(), sendParams())
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	return err
}

// A rate limit is transient and must not be confused with a used-up quota:
// one is retried after Retry-After, the other can never succeed by retrying.
func TestRateLimitIsRetryable(t *testing.T) {
	err := sendAgainst(t, tooManyRequests(ErrorCodeRateLimitExceeded, "7"))

	if !IsRateLimited(err) {
		t.Fatalf("expected a rate limit, got %v", err)
	}
	if IsQuotaExceeded(err) {
		t.Error("a rate limit must not read as a used-up quota")
	}
	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.RetryAfter != 7 {
		t.Errorf("expected RetryAfter 7, got %d", apiErr.RetryAfter)
	}
}

func TestQuotaExceededIsNotRetryable(t *testing.T) {
	for _, code := range []string{
		ErrorCodeQuotaExceeded,
		ErrorCodeDailyQuotaExceeded,
		ErrorCodeHourlyQuotaExceeded,
	} {
		t.Run(code, func(t *testing.T) {
			err := sendAgainst(t, tooManyRequests(code, ""))

			if !IsQuotaExceeded(err) {
				t.Fatalf("expected a used-up quota, got %v", err)
			}
			if IsRateLimited(err) {
				t.Error("a used-up quota must not read as a retryable rate limit")
			}
			if got := err.(*Error).ErrorCode; got != code {
				t.Errorf("expected error code %q to survive, got %q", code, got)
			}
		})
	}
}

// A 429 without an error code predates error codes; it was a rate limit.
func TestRateLimitWithoutErrorCode(t *testing.T) {
	err := sendAgainst(t, tooManyRequests("", "2"))

	if !IsRateLimited(err) {
		t.Errorf("expected a code-less 429 to read as a rate limit, got %v", err)
	}
	if IsQuotaExceeded(err) {
		t.Error("a code-less 429 must not read as a used-up quota")
	}
}

// An unknown 429 code might be a new quota; claiming it is retryable would
// invite exactly the retry loop the split exists to prevent.
func TestUnknownTooManyRequestsCodeIsNeither(t *testing.T) {
	err := sendAgainst(t, tooManyRequests("monthly_contacts_exceeded", ""))

	if IsRateLimited(err) || IsQuotaExceeded(err) {
		t.Errorf("an unknown 429 code must be neither helper, got %v", err)
	}
}

// The codes are only meaningful on a 429; the same string on another status is
// not a rate limit or a quota.
func TestRateLimitHelpersRequireTooManyRequests(t *testing.T) {
	cases := []*Error{
		{StatusCode: http.StatusBadRequest, ErrorCode: ErrorCodeRateLimitExceeded},
		{StatusCode: http.StatusServiceUnavailable},
		{StatusCode: http.StatusForbidden, ErrorCode: ErrorCodeQuotaExceeded},
	}
	for _, e := range cases {
		if IsRateLimited(e) {
			t.Errorf("IsRateLimited(%v) = true, want false", e)
		}
		if IsQuotaExceeded(e) {
			t.Errorf("IsQuotaExceeded(%v) = true, want false", e)
		}
	}
}

func TestRateLimitHelpersOnNonAPIErrors(t *testing.T) {
	for _, err := range []error{
		nil,
		errors.New("network down"),
		fmt.Errorf("wrapped: %w", errors.New("x")),
	} {
		if IsRateLimited(err) || IsQuotaExceeded(err) {
			t.Errorf("non-API error %v must be neither", err)
		}
	}
}

// A 429 is still neither of the other families.
func TestTooManyRequestsIsNotOtherErrors(t *testing.T) {
	err := sendAgainst(t, tooManyRequests(ErrorCodeRateLimitExceeded, "1"))

	if IsNotFound(err) || IsValidationError(err) || IsUnauthorized(err) || IsConflict(err) {
		t.Errorf("a 429 must not match another predicate, got %v", err)
	}
}

// A 429 from something in front of the API may lack Retry-After. It is still a
// rate limit, and RetryAfter is 0 - the doc tells callers to back off themselves.
func TestRateLimitWithoutRetryAfter(t *testing.T) {
	err := sendAgainst(t, tooManyRequests(ErrorCodeRateLimitExceeded, ""))

	if !IsRateLimited(err) {
		t.Fatalf("expected a rate limit, got %v", err)
	}
	if got := err.(*Error).RetryAfter; got != 0 {
		t.Errorf("expected RetryAfter 0 without the header, got %d", got)
	}
}
