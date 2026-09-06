package domain

import (
	"errors"
	"testing"
	"time"
)

func TestJournalEntryRequiresExactIntegerBalance(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	j := JournalEntry{ID: "J1", Description: "test", PostedAt: now, Lines: []JournalLine{{Position: 0, AccountID: "cash", DebitRial: 100}, {Position: 1, AccountID: "equity", CreditRial: 99}}}
	if !errors.Is(j.Validate(), ErrJournalUnbalanced) {
		t.Fatalf("unbalanced journal error=%v", j.Validate())
	}
	j.Lines[1].CreditRial = 100
	if err := j.Validate(); err != nil {
		t.Fatal(err)
	}
	r := j.Reversal("J2", "reverse:test", "reverse", now)
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	if r.Lines[0].CreditRial != 100 || r.Lines[1].DebitRial != 100 {
		t.Fatalf("reversal did not invert lines: %+v", r.Lines)
	}
}
