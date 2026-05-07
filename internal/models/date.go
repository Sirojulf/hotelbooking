package models

import (
	"fmt"
	"strings"
	"time"
)

// Date wraps time.Time to marshal/unmarshal as "YYYY-MM-DD" for Supabase date columns.
type Date struct {
	time.Time
}

func NewDate(t time.Time) Date {
	return Date{t}
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d.Time = t
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time = t
		return nil
	}
	return fmt.Errorf("cannot parse %q as date", s)
}
