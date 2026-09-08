package demo

import (
	"context"
	"testing"
)

func TestGenerateIsDeterministicAndInterconnected(t *testing.T) {
	ctx := context.Background()
	a, err := Generate(ctx, Options{Root: t.TempDir() + "/demo-a", Seed: DefaultSeed, ReferenceDate: DefaultReferenceDate})
	if err != nil {
		t.Fatal(err)
	}
	b, err := Generate(ctx, Options{Root: t.TempDir() + "/demo-b", Seed: DefaultSeed, ReferenceDate: DefaultReferenceDate})
	if err != nil {
		t.Fatal(err)
	}
	if a.Seed != b.Seed || a.Reference != b.Reference || a.DataDigest != b.DataDigest {
		t.Fatalf("same seed/reference produced different summaries: %#v %#v", a, b)
	}
	want := map[string]int{
		"customers": 4, "suppliers": 3, "materials": 4, "services": 3, "machines": 3,
		"purchases": 3, "orders": 4, "production": 2, "invoices": 2,
		"payments": 2, "expenses": 1, "transfers": 1, "checks": 2, "loans": 1,
		"owners": 2, "periods": 1,
	}
	for name, count := range want {
		if a.Counts[name] != count {
			t.Fatalf("%s count=%d, want %d; all counts=%v", name, a.Counts[name], count, a.Counts)
		}
	}
	if _, err := Generate(ctx, Options{Root: a.Root, Seed: DefaultSeed, ReferenceDate: DefaultReferenceDate}); err == nil {
		t.Fatal("generator accepted a marked existing demo root without reset")
	}
}
