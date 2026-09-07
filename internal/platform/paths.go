package platform

import (
	"errors"
	"os"
	"path/filepath"
)

const ApplicationVersion = "0.1.0"
const CurrentSchemaVersion = 13

type DataPaths struct {
	Root, Database, Attachments, Backups string
}

func ResolveDataPaths(appName string) (DataPaths, error) {
	if appName == "" {
		return DataPaths{}, errors.New("application name is required")
	}
	root := os.Getenv("ATROPATEN_DATA_DIR")
	if root == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return DataPaths{}, err
		}
		root = filepath.Join(base, appName)
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return DataPaths{}, err
	}
	return DataPaths{Root: root, Database: filepath.Join(root, "atropaten.db"), Attachments: filepath.Join(root, "attachments"), Backups: filepath.Join(root, "backups")}, nil
}

func (p DataPaths) Ensure() error {
	for _, dir := range []string{p.Root, p.Attachments, p.Backups} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	return nil
}
