package pagination

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/tangelo-labs/go-domainkit"
)

// ContinuationToken models a general purpose Continuation Token.
type ContinuationToken struct {
	// ID is the identity (primary key) of the last element in a concrete page.
	// This is necessary to distinguish between elements with the same timestamp.
	ID domain.ID

	// Timestamp is the time mark of the last element of a concrete page. Usually corresponds to a column like `created`
	// or `modified`. It's used to know where the next page should start.
	Timestamp time.Time
}

// FromString rebuilds a token from the given string, it's expected to be a string returned by the
// ContinuationToken.String() method.
func (ct *ContinuationToken) FromString(s string) {
	if s == "" {
		return
	}

	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return
	}

	if err := json.Unmarshal(b, &ct); err != nil {
		return
	}
}

// IsZero Whether this token is an empty token.
func (ct ContinuationToken) IsZero() bool {
	return ct.Timestamp.IsZero() && ct.ID.IsEmpty()
}

// Equals whether this token is equals to another.
func (ct ContinuationToken) Equals(other ContinuationToken) bool {
	return ct.ID.Equals(other.ID) && ct.Timestamp.Unix() == other.Timestamp.Unix()
}

// String returns a string representation of this token.
func (ct ContinuationToken) String() string {
	if ct.Timestamp.IsZero() && ct.ID.IsEmpty() {
		return ""
	}

	b, err := json.Marshal(ct)
	if err != nil {
		return ""
	}

	return base64.RawURLEncoding.EncodeToString(b)
}
