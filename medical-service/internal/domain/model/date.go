package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

type Date string

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return "", fmt.Errorf("date must be YYYY-MM-DD: %q", s)
	}
	return Date(t.Format(dateLayout)), nil
}

func DateOf(t time.Time) Date { return Date(t.Format(dateLayout)) }

func (d Date) AddDays(n int) Date {
	t, err := time.Parse(dateLayout, string(d))
	if err != nil {
		return d
	}
	return DateOf(t.AddDate(0, 0, n))
}

func (d Date) Value() (driver.Value, error) { return string(d), nil }

func (d *Date) Scan(src any) error {
	switch v := src.(type) {
	case time.Time:
		*d = DateOf(v)
	case []byte:
		*d = Date(v[:min(len(v), len(dateLayout))])
	case string:
		*d = Date(v[:min(len(v), len(dateLayout))])
	case nil:
		*d = ""
	default:
		return fmt.Errorf("cannot scan %T into Date", src)
	}
	return nil
}
