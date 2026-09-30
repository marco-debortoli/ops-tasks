package store

import (
	"encoding/json"
	"strings"
	"time"
)

// Opt is a JSON field for partial updates: absent (Set=false), null (Set, !Valid) or a value.
type Opt[T any] struct {
	Set   bool
	Valid bool
	V     T
}

func (o *Opt[T]) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Valid = false
		return nil
	}
	o.Valid = true
	return json.Unmarshal(b, &o.V)
}

// Ptr returns the value as a pointer, nil when null.
func (o Opt[T]) Ptr() *T {
	if !o.Valid {
		return nil
	}
	v := o.V
	return &v
}

func cleanName(field, s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", invalid("%s can't be empty", field)
	}
	return s, nil
}

// cleanDate validates an optional YYYY-MM-DD date.
func cleanDate(field string, o Opt[string]) (*string, error) {
	if !o.Valid {
		return nil, nil
	}
	d := strings.TrimSpace(o.V)
	if _, err := time.Parse(time.DateOnly, d); err != nil {
		return nil, invalid("%s must be a YYYY-MM-DD date", field)
	}
	return &d, nil
}

func samePtr[T comparable](a, b *T) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
