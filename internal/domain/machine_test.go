package domain

import (
	"testing"
	"time"
)

func TestMachineValidation(t *testing.T) {
	now := time.Date(2024, 8, 12, 7, 0, 0, 0, time.UTC)
	if _, err := NewMachine("MAC-1", MachineDraft{Name: "Printer", RateBasis: RatePerHour, RateRial: 100}, now); err != nil {
		t.Fatal(err)
	}
	if _, err := NewMachine("MAC-1", MachineDraft{Name: "Printer", RateBasis: "week", RateRial: 100}, now); err == nil {
		t.Fatal("unsupported rate basis accepted")
	}
	if _, err := NewMachine("MAC-1", MachineDraft{Name: "Printer", RateBasis: RatePerHour, RateRial: -1}, now); err == nil {
		t.Fatal("negative rate accepted")
	}
}
