package platform

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const BackupFormatVersion = 1

type BackupFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type BackupManifest struct {
	FormatVersion      int          `json:"formatVersion"`
	ApplicationVersion string       `json:"applicationVersion"`
	SchemaVersion      int          `json:"schemaVersion"`
	CreatedAt          string       `json:"createdAt"`
	Database           string       `json:"database"`
	Files              []BackupFile `json:"files"`
}
type BackupInfo struct {
	Path               string `json:"path"`
	CreatedAt          string `json:"createdAt"`
	ApplicationVersion string `json:"applicationVersion"`
	SchemaVersion      int    `json:"schemaVersion"`
	SizeBytes          int64  `json:"sizeBytes"`
	ManagedFileCount   int    `json:"managedFileCount"`
}

type BackupRepository interface {
	SnapshotTo(context.Context, string) error
	SchemaVersion(context.Context) (int, error)
	ManagedFilePaths(context.Context) ([]string, error)
}
type DatabaseValidator func(string) error
type BackupService struct {
	paths      DataPaths
	repository BackupRepository
	validate   DatabaseValidator
}

func NewBackupService(paths DataPaths, repository BackupRepository, validate DatabaseValidator) *BackupService {
	return &BackupService{paths: paths, repository: repository, validate: validate}
}
func (s *BackupService) SetRepository(repository BackupRepository) { s.repository = repository }
func (s *BackupService) Paths() DataPaths                          { return s.paths }

func (s *BackupService) Create(ctx context.Context) (BackupInfo, error) {
	if s.repository == nil {
		return BackupInfo{}, errors.New("backup repository is not configured")
	}
	if s.validate == nil {
		return BackupInfo{}, errors.New("database validator is not configured")
	}
	if err := s.paths.Ensure(); err != nil {
		return BackupInfo{}, err
	}
	tmp, err := os.MkdirTemp(s.paths.Backups, ".backup-")
	if err != nil {
		return BackupInfo{}, err
	}
	defer os.RemoveAll(tmp)
	dbPath := filepath.Join(tmp, "database.sqlite")
	if err = s.repository.SnapshotTo(ctx, dbPath); err != nil {
		return BackupInfo{}, fmt.Errorf("snapshot database: %w", err)
	}
	if err = s.validate(dbPath); err != nil {
		return BackupInfo{}, fmt.Errorf("validate snapshot: %w", err)
	}
	paths, err := s.repository.ManagedFilePaths(ctx)
	if err != nil {
		return BackupInfo{}, err
	}
	manifest := BackupManifest{FormatVersion: BackupFormatVersion, ApplicationVersion: ApplicationVersion, Database: "database.sqlite", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), Files: []BackupFile{}}
	for _, source := range paths {
		if _, err = os.Stat(source); err != nil {
			return BackupInfo{}, fmt.Errorf("managed file %q: %w", source, err)
		}
		rel, ok := s.managedRelative(source)
		if !ok {
			continue
		}
		target := filepath.Join(tmp, rel)
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return BackupInfo{}, err
		}
		if err = copyFile(source, target); err != nil {
			return BackupInfo{}, err
		}
		sum, size, err := fileDigest(target)
		if err != nil {
			return BackupInfo{}, err
		}
		manifest.Files = append(manifest.Files, BackupFile{Path: filepath.ToSlash(rel), SHA256: sum, Size: size})
	}
	var schema int
	schema, err = s.repository.SchemaVersion(ctx)
	if err != nil {
		return BackupInfo{}, err
	}
	manifest.SchemaVersion = schema
	archiveName := fmt.Sprintf("atropaten-%s.zip", time.Now().UTC().Format("20060102-150405.000000000"))
	tmpArchive := filepath.Join(s.paths.Backups, "."+archiveName+".tmp")
	if err = writeArchive(tmpArchive, tmp, manifest); err != nil {
		_ = os.Remove(tmpArchive)
		return BackupInfo{}, err
	}
	if err = s.verifyArchive(tmpArchive); err != nil {
		_ = os.Remove(tmpArchive)
		return BackupInfo{}, err
	}
	final := filepath.Join(s.paths.Backups, archiveName)
	if err = os.Rename(tmpArchive, final); err != nil {
		_ = os.Remove(tmpArchive)
		return BackupInfo{}, err
	}
	info := BackupInfo{Path: final, CreatedAt: manifest.CreatedAt, ApplicationVersion: manifest.ApplicationVersion, SchemaVersion: manifest.SchemaVersion, ManagedFileCount: len(manifest.Files)}
	if stat, e := os.Stat(final); e == nil {
		info.SizeBytes = stat.Size()
	}
	return info, nil
}

func (s *BackupService) Verify(path string) (BackupInfo, error) {
	if err := s.verifyArchive(path); err != nil {
		return BackupInfo{}, err
	}
	return s.archiveInfo(path)
}

func (s *BackupService) LastBackup() (BackupInfo, error) {
	entries, err := os.ReadDir(s.paths.Backups)
	if err != nil {
		return BackupInfo{}, err
	}
	var newest os.DirEntry
	var newestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return BackupInfo{}, err
		}
		if newest == nil || info.ModTime().After(newestTime) {
			newest, newestTime = entry, info.ModTime()
		}
	}
	if newest == nil {
		return BackupInfo{}, errors.New("no backups found")
	}
	return s.Verify(filepath.Join(s.paths.Backups, newest.Name()))
}

func (s *BackupService) verifyArchive(path string) error {
	if path == "" {
		return errors.New("backup path is required")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	z, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return fmt.Errorf("invalid backup archive: %w", err)
	}
	var manifest *BackupManifest
	entries := map[string]*zip.File{}
	for _, entry := range z.File {
		if !safeArchivePath(entry.Name) {
			return fmt.Errorf("unsafe archive path %q", entry.Name)
		}
		entries[entry.Name] = entry
		if entry.Name == "manifest.json" {
			var value BackupManifest
			if err := readJSON(entry, &value); err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}
			manifest = &value
		}
	}
	if manifest == nil {
		return errors.New("backup manifest is missing")
	}
	if manifest.FormatVersion != BackupFormatVersion {
		return fmt.Errorf("unsupported backup format version %d", manifest.FormatVersion)
	}
	if manifest.SchemaVersion < 1 || manifest.SchemaVersion > CurrentSchemaVersion {
		return fmt.Errorf("unsupported schema version %d", manifest.SchemaVersion)
	}
	if strings.TrimSpace(manifest.ApplicationVersion) == "" {
		return errors.New("backup application version is missing")
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt); err != nil {
		return fmt.Errorf("backup creation timestamp is invalid: %w", err)
	}
	if manifest.Database != "database.sqlite" {
		return errors.New("manifest database entry is invalid")
	}
	dbEntry := entries[manifest.Database]
	if dbEntry == nil {
		return errors.New("database snapshot is missing")
	}
	tmp, err := os.MkdirTemp("", "atropaten-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	dbPath := filepath.Join(tmp, "database.sqlite")
	if err = extractEntry(dbEntry, dbPath); err != nil {
		return err
	}
	if s.validate == nil {
		return errors.New("database validator is not configured")
	}
	if err = s.validate(dbPath); err != nil {
		return fmt.Errorf("validate database: %w", err)
	}
	seenFiles := map[string]bool{}
	for _, file := range manifest.Files {
		if !safeArchivePath(file.Path) || !strings.HasPrefix(file.Path, "attachments/") {
			return fmt.Errorf("managed file path is invalid: %q", file.Path)
		}
		if seenFiles[file.Path] {
			return fmt.Errorf("duplicate managed file: %s", file.Path)
		}
		seenFiles[file.Path] = true
		entry := entries[file.Path]
		if entry == nil {
			return fmt.Errorf("managed file is missing: %s", file.Path)
		}
		dest := filepath.Join(tmp, filepath.FromSlash(file.Path))
		if err = os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
			return err
		}
		if err = extractEntry(entry, dest); err != nil {
			return err
		}
		sum, size, err := fileDigest(dest)
		if err != nil {
			return err
		}
		if sum != file.SHA256 || size != file.Size {
			return fmt.Errorf("checksum mismatch for %s", file.Path)
		}
	}
	return nil
}

// Restore atomically swaps only the database and managed attachment tree. The callback closes/reopens the live store; on reopen failure both paths are rolled back.
func (s *BackupService) Restore(ctx context.Context, path string, closeLive func() error, reopen func() error) (BackupInfo, error) {
	if err := s.verifyArchive(path); err != nil {
		return BackupInfo{}, err
	}
	tmp, err := os.MkdirTemp(s.paths.Root, ".restore-")
	if err != nil {
		return BackupInfo{}, err
	}
	defer os.RemoveAll(tmp)
	if err = extractArchive(path, tmp); err != nil {
		return BackupInfo{}, err
	}
	if err = os.MkdirAll(filepath.Join(tmp, "attachments"), 0o700); err != nil {
		return BackupInfo{}, err
	}
	if err = closeLive(); err != nil {
		return BackupInfo{}, fmt.Errorf("close live database: %w", err)
	}
	rollbackRoot := filepath.Join(s.paths.Root, ".rollback-"+time.Now().UTC().Format("20060102-150405.000000000"))
	if err = os.MkdirAll(rollbackRoot, 0o700); err != nil {
		_ = reopen()
		return BackupInfo{}, err
	}
	oldDB := filepath.Join(rollbackRoot, "atropaten.db")
	oldAttachments := filepath.Join(rollbackRoot, "attachments")
	movedDB, movedAttachments, installedDB, installedAttachments := false, false, false, false
	rollback := func() {
		if installedAttachments {
			_ = os.RemoveAll(s.paths.Attachments)
		}
		if installedDB {
			_ = os.Remove(s.paths.Database)
		}
		if movedAttachments {
			_ = os.Rename(oldAttachments, s.paths.Attachments)
		}
		if movedDB {
			_ = os.Rename(oldDB, s.paths.Database)
		}
		_ = os.RemoveAll(rollbackRoot)
	}
	move := func(source, target string) (bool, error) {
		err := os.Rename(source, target)
		if os.IsNotExist(err) {
			return false, nil
		}
		return err == nil, err
	}
	if movedDB, err = move(s.paths.Database, oldDB); err != nil {
		rollback()
		_ = reopen()
		return BackupInfo{}, err
	}
	if movedAttachments, err = move(s.paths.Attachments, oldAttachments); err != nil {
		rollback()
		_ = reopen()
		return BackupInfo{}, err
	}
	if err = os.Rename(filepath.Join(tmp, "database.sqlite"), s.paths.Database); err != nil {
		rollback()
		_ = reopen()
		return BackupInfo{}, err
	}
	installedDB = true
	if err = os.Rename(filepath.Join(tmp, "attachments"), s.paths.Attachments); err != nil {
		rollback()
		_ = reopen()
		return BackupInfo{}, err
	}
	installedAttachments = true
	if err = reopen(); err != nil {
		_ = closeLive()
		rollback()
		if reopenErr := reopen(); reopenErr != nil {
			return BackupInfo{}, fmt.Errorf("restore reopen failed and rollback reopen failed: %v; %v", err, reopenErr)
		}
		return BackupInfo{}, fmt.Errorf("restore reopen failed; original data restored: %w", err)
	}
	_ = os.RemoveAll(rollbackRoot)
	return s.archiveInfo(path)
}

func (s *BackupService) managedRelative(source string) (string, bool) {
	root, _ := filepath.Abs(s.paths.Attachments)
	abs, err := filepath.Abs(source)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", false
	}
	return filepath.Join("attachments", rel), true
}
func (s *BackupService) archiveInfo(path string) (BackupInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return BackupInfo{}, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return BackupInfo{}, err
	}
	z, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return BackupInfo{}, err
	}
	for _, entry := range z.File {
		if entry.Name != "manifest.json" {
			continue
		}
		var m BackupManifest
		if err = readJSON(entry, &m); err != nil {
			return BackupInfo{}, err
		}
		return BackupInfo{Path: path, CreatedAt: m.CreatedAt, ApplicationVersion: m.ApplicationVersion, SchemaVersion: m.SchemaVersion, SizeBytes: stat.Size(), ManagedFileCount: len(m.Files)}, nil
	}
	return BackupInfo{}, errors.New("backup manifest is missing")
}
func writeArchive(path, root string, m BackupManifest) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	z := zip.NewWriter(f)
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		_ = f.Close()
		return err
	}
	w, err := z.Create("manifest.json")
	if err != nil {
		_ = f.Close()
		return err
	}
	if _, err = w.Write(raw); err != nil {
		_ = f.Close()
		return err
	}
	if err = walkArchive(z, root, root); err != nil {
		_ = z.Close()
		_ = f.Close()
		return err
	}
	if err = z.Close(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
func walkArchive(z *zip.Writer, root, current string) error {
	return filepath.Walk(current, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "manifest.json" {
			return nil
		}
		w, e := z.Create(filepath.ToSlash(rel))
		if e != nil {
			return e
		}
		r, e := os.Open(path)
		if e != nil {
			return e
		}
		defer r.Close()
		_, e = io.Copy(w, r)
		return e
	})
}
func extractArchive(path, dest string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	stat, _ := f.Stat()
	z, err := zip.NewReader(f, stat.Size())
	if err != nil {
		return err
	}
	for _, entry := range z.File {
		if entry.Name == "manifest.json" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(entry.Name))
		if !safeArchivePath(entry.Name) {
			return fmt.Errorf("unsafe archive path %q", entry.Name)
		}
		if info := entry.FileInfo(); info.IsDir() {
			if err = os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err = extractEntry(entry, target); err != nil {
			return err
		}
	}
	return nil
}
func extractEntry(entry *zip.File, target string) error {
	r, err := entry.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	f, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
func readJSON(entry *zip.File, v any) error {
	r, err := entry.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	return json.NewDecoder(r).Decode(v)
}
func copyFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(w, r)
	closeErr := w.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
func fileDigest(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), n, err
}
func safeArchivePath(name string) bool {
	if name == "" || filepath.IsAbs(filepath.FromSlash(name)) || strings.Contains(name, "\\") || (len(name) >= 2 && name[1] == ':') {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	return clean == name && clean != "." && !strings.HasPrefix(clean, "../") && !strings.Contains(clean, "/../")
}
