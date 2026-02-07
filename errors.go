package lettr

import (
	"encoding/json"
	"fmt"
	"net/http"
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

	// Errors contains field-level validation errors (for 422 responses).
	Errors map[string][]string `json:"errors,omitempty"`
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

	return apiErr
}
