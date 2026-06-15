package collection

import (
	"time"
)

// TellicoDate represents a date in the Tellico format.
type TellicoDate struct {
	Literal string
	YYYY    int
	MM      int
	DD      int
}

// NewTellicoDate creates a new TellicoDate from a literal string.
func NewTellicoDate(literal string) *TellicoDate {
	if literal == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", literal); err == nil {
		return &TellicoDate{
			YYYY: t.Year(),
			MM:   int(t.Month()),
			DD:   t.Day(),
		}
	}
	return &TellicoDate{
		Literal: literal,
	}
}
