package lettr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Error represents an error returned by the Lettr API.
type Error struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int `json:"-"`

	// Message is a human-readable error message.
	Message string `json:"message"`

	// ErrorCode is a machine-readable error code (e.g. "validation_error", "not_found").
	ErrorCode string `json:"error_code,omitempty"`

	// RetryAfter is the Retry-After header in seconds, when the API sent one.
	//
	// It is what separates the retryable failures from the permanent ones:
	// idempotency_in_progress carries it and should be retried with the same
	// key, while idempotency_key_conflict does not and never will succeed.
	RetryAfter int `json:"-"`

	// Errors contains field-level validation errors (for 422 responses).
	Errors map[string][]string `json:"errors,omitempty"`
}

// UnmarshalJSON tolerates a PHP/Laravel serialization quirk where an empty
// associative array can come over the wire as `[]` instead of `{}`. For
// Error.Errors that means a 422 response with no field-level errors may carry
// `"errors": []`; without this tolerance the standard library would fail with
// "cannot unmarshal array into Go struct field … of type map[string][]string"
// and the SDK would surface a confusing decode error instead of the real 422.
func (e *Error) UnmarshalJSON(data []byte) error {
	// alias prevents infinite recursion into this method when we Unmarshal
	// back through the embedded struct.
	type alias Error
	aux := struct {
		*alias
		Errors json.RawMessage `json:"errors,omitempty"`
	}{
		alias: (*alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Errors) == 0 {
		return nil
	}
	trimmed := bytes.TrimSpace(aux.Errors)
	if bytes.Equal(trimmed, []byte("[]")) || bytes.Equal(trimmed, []byte("null")) {
		// No field-level errors — leave e.Errors nil.
		return nil
	}
	return json.Unmarshal(aux.Errors, &e.Errors)
}

// Error implements the error interface.
func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("lettr: %d %s", e.StatusCode, e.Message))
	if e.ErrorCode != "" {
		sb.WriteString(fmt.Sprintf(" (code: %s)", e.ErrorCode))
	}
	if len(e.Errors) > 0 {
		for field, msgs := range e.Errors {
			for _, msg := range msgs {
				sb.WriteString(fmt.Sprintf("; %s: %s", field, msg))
			}
		}
	}
	return sb.String()
}

// IsNotFound returns true if the error is a 404 Not Found error.
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusNotFound
	}
	return false
}

// IsValidationError returns true if the error is a 422 Validation Error.
func IsValidationError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusUnprocessableEntity
	}
	return false
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func IsUnauthorized(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusUnauthorized
	}
	return false
}

// ErrorCodeIdempotencyKeyConflict is sent when an Idempotency-Key was already
// used with a different request payload.
const ErrorCodeIdempotencyKeyConflict = "idempotency_key_conflict"

// ErrorCodeIdempotencyInProgress is sent when the original request for an
// Idempotency-Key is still being processed.
const ErrorCodeIdempotencyInProgress = "idempotency_in_progress"

// IsIdempotencyConflict reports whether err is the 409 for an Idempotency-Key
// reused with a different payload.
//
// Never retry this. Two different emails were sent under one key, which is a
// bug on the caller's side; the same request will fail identically forever.
// Use a key that is unique per logical send, or send the payload the key was
// first used with.
func IsIdempotencyConflict(err error) bool {
	e, ok := err.(*Error)
	if !ok || e.StatusCode != http.StatusConflict {
		return false
	}
	return e.ErrorCode == ErrorCodeIdempotencyKeyConflict
}

// IsIdempotencyInProgress reports whether err is the 409 for a send whose
// original request is still running.
//
// Unlike IsIdempotencyConflict this one is retryable, and must be retried with
// the same key - a fresh key would send a second email. Wait Error.RetryAfter
// seconds first.
func IsIdempotencyInProgress(err error) bool {
	e, ok := err.(*Error)
	if !ok || e.StatusCode != http.StatusConflict {
		return false
	}
	return e.ErrorCode == ErrorCodeIdempotencyInProgress
}

// ErrorCodeResourceAlreadyExists is the machine-readable code the API sends
// when a create collides with an existing resource.
const ErrorCodeResourceAlreadyExists = "resource_already_exists"

// IsConflict returns true if the error is a 409 Conflict error.
func IsConflict(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.StatusCode == http.StatusConflict
	}
	return false
}

// IsContactAlreadyExists returns true if the error is the 409 that
// AudienceContactService.Create returns when the email is already in the team's
// audience.
//
//	_, err := client.Audience.Contacts.Create(ctx, &lettr.CreateAudienceContactRequest{Email: email})
//	if lettr.IsContactAlreadyExists(err) {
//		// Client-correctable: update the existing contact instead.
//	}
//
// This is not a retryable failure. The API used to let a duplicate escape as an
// HTTP 500 with the misleading "send_error" code (it names email delivery,
// which is not involved here); a retry-on-5xx policy would retry it pointlessly.
// It is now a 409, and a 409 here must not be retried.
//
// A 409 that carries no error code also counts, so the check still works
// against an API deployment that predates the change.
func IsContactAlreadyExists(err error) bool {
	e, ok := err.(*Error)
	if !ok || e.StatusCode != http.StatusConflict {
		return false
	}
	return e.ErrorCode == "" || e.ErrorCode == ErrorCodeResourceAlreadyExists
}

// parseError reads the response body and constructs an *Error.
func parseError(resp *http.Response) error {
	apiErr := &Error{
		StatusCode: resp.StatusCode,
	}

	if resp.Body == nil {
		apiErr.Message = http.StatusText(resp.StatusCode)
		return apiErr
	}

	if err := json.NewDecoder(resp.Body).Decode(apiErr); err != nil {
		apiErr.Message = http.StatusText(resp.StatusCode)
	}

	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(resp.StatusCode)
	}

	if seconds, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && seconds > 0 {
		apiErr.RetryAfter = seconds
	}

	return apiErr
}
