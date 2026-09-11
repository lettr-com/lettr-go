package lettr

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func sendParams() *SendEmailRequest {
	return &SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Hello",
		Html:    "<p>Hi</p>",
	}
}

func writeAccepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Email queued for delivery.",
		"data":    map[string]any{"request_id": "req-1", "accepted": 1, "rejected": 0},
	})
}

func TestSendPutsTheKeyInTheHeader(t *testing.T) {
	var gotKey string
	var gotBody map[string]any

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeAccepted(w)
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), sendParams(),
		WithIdempotencyKey("order-12345"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotKey != "order-12345" {
		t.Errorf("expected the key in the header, got %q", gotKey)
	}
	if _, present := gotBody["idempotency_key"]; present {
		t.Error("the key must travel as a header, never in the body")
	}
}

// The compatibility guarantee: an existing two-argument call sends the exact
// request it sent before this existed.
func TestSendWithoutAKeySendsNoHeader(t *testing.T) {
	var gotKey string

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("Idempotency-Key")
		writeAccepted(w)
	})
	defer server.Close()

	resp, err := client.Emails.Send(context.Background(), sendParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotKey != "" {
		t.Errorf("expected no Idempotency-Key header, got %q", gotKey)
	}
	if resp.Data.Replayed {
		t.Error("a send without a key can never be a replay")
	}
}

// Validated locally, so a bad key costs no round trip and no 422.
func TestSendRejectsAMalformedKeyBeforeSending(t *testing.T) {
	called := false

	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		writeAccepted(w)
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), sendParams(),
		WithIdempotencyKey("order 12345"))

	if err == nil {
		t.Fatal("expected an error for a malformed key")
	}
	if called {
		t.Error("no request should have been made")
	}
}

func TestIsValidIdempotencyKey(t *testing.T) {
	valid := []string{"order-confirmation-12345", "a.b_c-1", "a"}
	invalid := []string{"", "order 123", "order/123", "order:123", "order-č"}

	for _, key := range valid {
		if !IsValidIdempotencyKey(key) {
			t.Errorf("expected %q to be valid", key)
		}
	}
	for _, key := range invalid {
		if IsValidIdempotencyKey(key) {
			t.Errorf("expected %q to be invalid", key)
		}
	}

	long := make([]byte, 255)
	for i := range long {
		long[i] = 'a'
	}
	if !IsValidIdempotencyKey(string(long)) {
		t.Error("expected a 255-character key to be valid")
	}
	if IsValidIdempotencyKey(string(long) + "a") {
		t.Error("expected a 256-character key to be invalid")
	}
}

func TestSendMarksAReplay(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Idempotency-Replayed", "true")
		writeAccepted(w)
	})
	defer server.Close()

	resp, err := client.Emails.Send(context.Background(), sendParams(),
		WithIdempotencyKey("order-12345"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No second email went out, but nothing failed either.
	if !resp.Data.Replayed {
		t.Error("expected the send to be marked as replayed")
	}
	if resp.Data.Accepted != 1 {
		t.Errorf("a replay is still a success, got accepted=%d", resp.Data.Accepted)
	}
}

// The two 409s must not be conflated: one is safe to retry with the same key,
// the other will fail forever.
func TestIdempotencyInProgressIsRetryable(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "3")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":    "A request with this Idempotency-Key is still processing.",
			"error_code": "idempotency_in_progress",
		})
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), sendParams(),
		WithIdempotencyKey("order-12345"))

	if !IsIdempotencyInProgress(err) {
		t.Fatalf("expected an in-progress conflict, got %v", err)
	}
	if IsIdempotencyConflict(err) {
		t.Error("an in-progress conflict must not read as a payload conflict")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.RetryAfter != 3 {
		t.Errorf("expected RetryAfter 3, got %d", apiErr.RetryAfter)
	}
}

func TestIdempotencyKeyConflictIsNotRetryable(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":    "This Idempotency-Key was already used with a different request payload.",
			"error_code": "idempotency_key_conflict",
		})
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), sendParams(),
		WithIdempotencyKey("order-12345"))

	if !IsIdempotencyConflict(err) {
		t.Fatalf("expected a payload conflict, got %v", err)
	}
	if IsIdempotencyInProgress(err) {
		t.Error("a payload conflict must not read as retryable")
	}

	apiErr, _ := err.(*Error)
	if apiErr.RetryAfter != 0 {
		t.Errorf("nothing to wait for, got RetryAfter=%d", apiErr.RetryAfter)
	}
}

func TestAnUnrelatedConflictIsNeitherIdempotencyError(t *testing.T) {
	client, server := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":    "A contact with this email already exists.",
			"error_code": "resource_already_exists",
		})
	})
	defer server.Close()

	_, err := client.Emails.Send(context.Background(), sendParams())

	if IsIdempotencyConflict(err) || IsIdempotencyInProgress(err) {
		t.Error("an unrelated 409 must not read as an idempotency failure")
	}
	if !IsConflict(err) {
		t.Error("expected it to still be a conflict")
	}
}
