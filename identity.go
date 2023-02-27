package domain

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// ULIDGenerator is the default ID generator function that uses ULID algorithm.
var ULIDGenerator = func() ID {
	return ID(ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String())
}

// UUIDGenerator is an ID generator function that uses UUID algorithm.
var UUIDGenerator = func() ID {
	return ID(uuid.New().String())
}

var idGenerator = ULIDGenerator

// SetIDGenerator sets the ID generator function. Call this function before
// using any of the ID generating functions, preferably in an init() function.
//
// By default, the ULID algorithm is used.
func SetIDGenerator(fn func() ID) {
	idGenerator = fn
}

// ID defines a globally unique identifier for an Entity.
type ID string

// NewID creates a new ID using an ULID generator.
func NewID() ID {
	return idGenerator()
}

// Equals returns true if this ID is equal to another.
func (id ID) Equals(other ID) bool {
	return id == other
}

// IsEmpty whether this ID is an empty ID or not (zero-valued).
func (id ID) IsEmpty() bool {
	return id == ""
}

// MarshalBinary encodes this ID as binary.
func (id ID) MarshalBinary() (data []byte, err error) {
	return []byte(id), nil
}

// UnmarshalBinary decodes this ID back from binary.
func (id *ID) UnmarshalBinary(data []byte) error {
	*id = ID(data)

	return nil
}

// MarshalJSON encodes this ID as JSON.
func (id ID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", id)), nil
}

// UnmarshalJSON decodes this ID back from JSON.
func (id *ID) UnmarshalJSON(bytes []byte) error {
	s := string(bytes)
	s = strings.Trim(s, "\"")
	*id = ID(s)

	return nil
}

// MarshalText marshals into a textual form.
func (id ID) MarshalText() (text []byte, err error) {
	return []byte(id), nil
}

// UnmarshalText unmarshal a textual representation of an ID.
func (id *ID) UnmarshalText(text []byte) error {
	*id = ID(text)

	return nil
}

// String returns a string representation of this ID.
func (id ID) String() string {
	return string(id)
}

// IDs a convenience definition for dealing with collection of IDs in memory.
type IDs []ID

// Contains whether this list contains the specified ID.
func (ids IDs) Contains(id ID) bool {
	for _, i := range ids {
		if i.Equals(id) {
			return true
		}
	}

	return false
}

// Subtract returns a new list in which the provided list is subtracted from the
// current list. Examples:
//
// * {} Subtract {} = {}.
// * {a,b,c,d,d} Subtract {b,d} = {a,c}.
// * {a,b,c,d} Subtract {} = {a,b,c,d}.
// * {a,b,c,d} Subtract {e} = {a,b,c,d}.
func (ids IDs) Subtract(others IDs) IDs {
	if len(others) == 0 {
		return ids
	}

	m := make(map[ID]bool, len(others))
	for i := range others {
		m[others[i]] = false
	}

	diff := make([]ID, 0)
	j := 0

	for i := range ids {
		if _, ok := m[ids[i]]; ok {
			continue
		}

		diff = append(diff, ids[i])
		j++
	}

	return diff
}
