package lettr

import (
	"fmt"
	"regexp"
)

// idempotencyKeyPattern is the format the API accepts: 1-255 characters of
// letters, digits, periods, underscores or hyphens.
var idempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,255}$`)

// IsValidIdempotencyKey reports whether a string is a usable idempotency key.
//
// Exported so callers deriving keys from their own ids - an order number, a job
// id - can check before sending rather than discovering it as a 422.
func IsValidIdempotencyKey(key string) bool {
	return idempotencyKeyPattern.MatchString(key)
}

// SendOption configures a single Send call.
type SendOption func(*sendConfig)

type sendConfig struct {
	idempotencyKey string
}

// WithIdempotencyKey sends the given Idempotency-Key with the request.
//
// Reuse the key when you retry and the API returns the original result instead
// of delivering a second email, with SendEmailData.Replayed set.
//
// You choose the key; the SDK never generates one. It only works if both
// attempts use the same value, and the SDK does not retry - one Send is one
// HTTP request - so the retry is yours, and only you know that two calls are
// the same logical send. A key generated inside Send would differ on every
// attempt and protect nothing while looking like it did.
//
// The key must be 1-255 characters of [A-Za-z0-9._-]; Send returns an error
// before making a request if it is not.
func WithIdempotencyKey(key string) SendOption {
	return func(c *sendConfig) {
		c.idempotencyKey = key
	}
}

func (c *sendConfig) validate() error {
	if c.idempotencyKey == "" {
		return nil
	}

	if !IsValidIdempotencyKey(c.idempotencyKey) {
		return fmt.Errorf(
			"lettr: invalid idempotency key %q: use 1 to 255 letters, digits, periods, underscores or hyphens",
			c.idempotencyKey,
		)
	}

	return nil
}
