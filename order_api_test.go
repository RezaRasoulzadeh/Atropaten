package main

import (
	"strings"
	"testing"
	"time"
)

func TestOrderInputPromisedAt(t *testing.T) {
	for _, tt := range []struct {
		name    string
		value   string
		unset   bool
		want    string
		wantErr bool
	}{
		{name: "unset", unset: true},
		{name: "empty"},
		{name: "date from picker", value: "2026-09-02", want: "2026-09-02T00:00:00Z"},
		{name: "timestamp", value: "2026-09-02T12:30:00Z", want: "2026-09-02T12:30:00Z"},
		{name: "timestamp with offset", value: "2026-09-02T12:30:00+03:30", want: "2026-09-02T09:00:00Z"},
		{name: "fractional seconds", value: "2026-09-02T12:30:00.123456789Z", want: "2026-09-02T12:30:00.123456789Z"},
		{name: "leap day", value: "2028-02-29", want: "2028-02-29T00:00:00Z"},
		{name: "invalid day", value: "2026-02-29", wantErr: true},
		{name: "invalid month", value: "2026-13-02", wantErr: true},
		{name: "malformed", value: "not-a-date", wantErr: true},
		{name: "timestamp without timezone", value: "2026-09-02T12:30:00", wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := OrderInput{}
			if !tt.unset {
				input.PromisedAt = &tt.value
			}
			got, err := orderInput(input)
			if tt.wantErr {
				if err == nil || !strings.HasPrefix(err.Error(), "promisedAt: invalid timestamp:") {
					t.Fatalf("expected promisedAt validation error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.want == "" {
				if got.PromisedAt != nil {
					t.Fatalf("expected no promised date, got %v", got.PromisedAt)
				}
				return
			}
			if got.PromisedAt == nil {
				t.Fatal("expected a promised date")
			}
			if actual := got.PromisedAt.UTC().Format(time.RFC3339Nano); actual != tt.want {
				t.Fatalf("promised date = %q, want %q", actual, tt.want)
			}
		})
	}
}
