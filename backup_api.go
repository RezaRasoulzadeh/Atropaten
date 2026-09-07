package main

import (
	"fmt"

	"Atropaten/internal/platform"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DataPathsDTO struct {
	Root               string `json:"root"`
	Database           string `json:"database"`
	Attachments        string `json:"attachments"`
	Backups            string `json:"backups"`
	ApplicationVersion string `json:"applicationVersion"`
	SchemaVersion      int    `json:"schemaVersion"`
}
type BackupInfoDTO struct {
	Path               string `json:"path"`
	CreatedAt          string `json:"createdAt"`
	ApplicationVersion string `json:"applicationVersion"`
	SchemaVersion      int    `json:"schemaVersion"`
	SizeBytes          int64  `json:"sizeBytes"`
	ManagedFileCount   int    `json:"managedFileCount"`
}

func (a *App) backupService() (*platform.BackupService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.backup == nil {
		return nil, fmt.Errorf("backup service is not initialized")
	}
	return a.backup, nil
}

func (a *App) GetDataPaths() (DataPathsDTO, error) {
	s, err := a.backupService()
	if err != nil {
		return DataPathsDTO{}, err
	}
	p := s.Paths()
	version, err := a.database.SchemaVersion(a.materialContext())
	if err != nil {
		return DataPathsDTO{}, err
	}
	return DataPathsDTO{Root: p.Root, Database: p.Database, Attachments: p.Attachments, Backups: p.Backups, ApplicationVersion: platform.ApplicationVersion, SchemaVersion: version}, nil
}
func (a *App) CreateBackup() (BackupInfoDTO, error) {
	s, err := a.backupService()
	if err != nil {
		return BackupInfoDTO{}, err
	}
	v, err := s.Create(a.materialContext())
	return backupInfoDTO(v), err
}
func (a *App) VerifyBackup(path string) (BackupInfoDTO, error) {
	s, err := a.backupService()
	if err != nil {
		return BackupInfoDTO{}, err
	}
	v, err := s.Verify(path)
	return backupInfoDTO(v), err
}
func (a *App) GetLastBackup() (BackupInfoDTO, error) {
	s, err := a.backupService()
	if err != nil {
		return BackupInfoDTO{}, err
	}
	v, err := s.LastBackup()
	return backupInfoDTO(v), err
}
func (a *App) RestoreBackup(path string) (BackupInfoDTO, error) {
	s, err := a.backupService()
	if err != nil {
		return BackupInfoDTO{}, err
	}
	v, err := s.Restore(a.materialContext(), path, a.closeForRestore, a.reopenAfterRestore)
	return backupInfoDTO(v), err
}
func (a *App) SelectBackupFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context is not initialized")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select Atropaten backup", Filters: []runtime.FileFilter{{DisplayName: "Atropaten backup", Pattern: "*.zip"}}})
}
func backupInfoDTO(v platform.BackupInfo) BackupInfoDTO {
	return BackupInfoDTO{Path: v.Path, CreatedAt: v.CreatedAt, ApplicationVersion: v.ApplicationVersion, SchemaVersion: v.SchemaVersion, SizeBytes: v.SizeBytes, ManagedFileCount: v.ManagedFileCount}
}
