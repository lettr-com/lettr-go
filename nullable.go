package lettr

import "encoding/json"

// NullString lets PATCH callers distinguish three intents for a nullable
// string field on the wire. Field names and zero-value semantics mirror
// database/sql's NullString: the zero value is null.
//
//   - nil *NullString                       → omit the field entirely
//                                             (server keeps the existing value)
//   - &NullString{}                         → send JSON null
//                                             (server clears the field)
//   - &NullString{Valid: true, String: "x"} → send the JSON string "x"
//                                             (use NewNullString as shorthand)
//
// Plain Go *string + ,omitempty cannot represent the three states because the
// json package treats a nil pointer the same as a missing field. NullString
// fills that gap for the audience PATCH endpoints where the API distinguishes
// "field omitted" (no change) from "field is null" (clear).
type NullString struct {
	String string
	// Valid indicates whether String is the value to send. When false, the
	// NullString marshals to JSON null regardless of String.
	Valid bool
}

// NewNullString returns a NullString that marshals to the given JSON string.
// It is shorthand for &NullString{Valid: true, String: s}.
func NewNullString(s string) *NullString {
	return &NullString{String: s, Valid: true}
}

// MarshalJSON implements json.Marshaler.
func (n *NullString) MarshalJSON() ([]byte, error) {
	if n == nil || !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}
