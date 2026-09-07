package demo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Atropaten/internal/platform"
)

const (
	MarkerFile = ".atropaten-demo.json"
	MarkerKind = "atropaten-development-demo"
)

type Marker struct {
	Kind          string `json:"kind"`
	Seed          int64  `json:"seed"`
	ReferenceDate string `json:"reference_date"`
	Root          string `json:"root"`
}

func NormalizeRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("demo root is required; refusing to use the normal application data directory")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve demo root: %w", err)
	}
	return filepath.Clean(abs), nil
}

func AssertIsolatedRoot(root string) error {
	root, err := NormalizeRoot(root)
	if err != nil {
		return err
	}
	production, err := platform.ResolveDataPaths("Atropaten")
	if err != nil {
		return fmt.Errorf("resolve production data directory: %w", err)
	}
	if root == filepath.Clean(production.Root) {
		return fmt.Errorf("refusing demo operation on production data directory %q", root)
	}
	if tempRoot, tempErr := filepath.Abs(os.TempDir()); tempErr == nil && root == filepath.Clean(tempRoot) {
		return fmt.Errorf("refusing to use the temporary directory itself; choose a dedicated child demo root")
	}
	if root == string(filepath.Separator) || root == filepath.Dir(root) {
		return fmt.Errorf("refusing unsafe demo root %q", root)
	}
	return nil
}

func ReadMarker(root string) (Marker, error) {
	if err := AssertIsolatedRoot(root); err != nil {
		return Marker{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, MarkerFile))
	if err != nil {
		return Marker{}, err
	}
	var marker Marker
	if err := json.Unmarshal(data, &marker); err != nil {
		return Marker{}, fmt.Errorf("read demo marker: %w", err)
	}
	clean, err := NormalizeRoot(root)
	if err != nil {
		return Marker{}, err
	}
	if marker.Kind != MarkerKind || filepath.Clean(marker.Root) != clean {
		return Marker{}, errors.New("demo marker does not match this root")
	}
	return marker, nil
}

func WriteMarker(root string, marker Marker) error {
	if err := AssertIsolatedRoot(root); err != nil {
		return err
	}
	clean, err := NormalizeRoot(root)
	if err != nil {
		return err
	}
	marker.Kind = MarkerKind
	marker.Root = clean
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(clean, MarkerFile), append(data, '\n'), 0o600)
}

func Reset(root string) error {
	marker, err := ReadMarker(root)
	if err != nil {
		return fmt.Errorf("reset requires a valid %s: %w", MarkerFile, err)
	}
	if marker.Kind != MarkerKind {
		return errors.New("refusing reset of an unmarked directory")
	}
	clean, err := NormalizeRoot(root)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(clean); err != nil {
		return fmt.Errorf("remove demo root: %w", err)
	}
	return nil
}
