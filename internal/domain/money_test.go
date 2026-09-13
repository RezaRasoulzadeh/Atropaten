package domain

import (
	"math"
	"strings"
	"testing"
)

func TestRoundMoneyUp(t *testing.T) {
	tests := []struct {
		name, wantErr      string
		amount, step, want int64
	}{
		{name: "step one", amount: 12001, step: 1, want: 12001},
		{name: "exact hundred", amount: 12100, step: 100, want: 12100},
		{name: "hundred", amount: 12001, step: 100, want: 12100},
		{name: "thousand", amount: 12001, step: 1000, want: 13000},
		{name: "thousand exact", amount: 13000, step: 1000, want: 13000},
		{name: "zero", amount: 0, step: 1000, want: 0},
		{name: "negative ceiling", amount: -12001, step: 1000, want: -12000},
		{name: "negative exact", amount: -12000, step: 1000, want: -12000},
		{name: "large divisible", amount: math.MaxInt64, step: 1, want: math.MaxInt64},
		{name: "large overflow", amount: math.MaxInt64, step: 2, wantErr: "exceeds"},
		{name: "invalid step", amount: 1, step: 0, wantErr: "positive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RoundMoneyUp(tt.amount, tt.step)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error=%v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %d/%v, want %d", got, err, tt.want)
			}
		})
	}
}
