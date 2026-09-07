package demo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResetRequiresOurMarker(t *testing.T) {
	root := t.TempDir()
	if err := Reset(root); err == nil {
		t.Fatal("Reset accepted an unmarked directory")
	}
	if err := WriteMarker(root, Marker{Seed: DefaultSeed, ReferenceDate: DefaultReferenceDate}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sentinel"), []byte("demo"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Reset(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("marked demo root still exists, stat error=%v", err)
	}
}

func TestAssertIsolatedRootRejectsEmptyAndProductionRoot(t *testing.T) {
	if err := AssertIsolatedRoot(""); err == nil {
		t.Fatal("empty root was accepted")
	}
	if err := AssertIsolatedRoot(filepath.Clean(filepath.Join(string(filepath.Separator), "tmp"))); err == nil {
		t.Fatal("unsafe broad root was accepted")
	}
}
