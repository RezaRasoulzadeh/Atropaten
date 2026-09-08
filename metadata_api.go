package main

import (
	"Atropaten/internal/application"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AttachmentDTO struct {
	ID        string `json:"id"`
	OwnerType string `json:"ownerType"`
	OwnerID   string `json:"ownerId"`
	FileName  string `json:"fileName"`
	Path      string `json:"path"`
	MIMEType  string `json:"mimeType"`
	Checksum  string `json:"checksum"`
	Category  string `json:"category"`
	Notes     string `json:"notes"`
	SizeBytes *int64 `json:"sizeBytes"`
	CreatedAt string `json:"createdAt"`
}
type AttachmentPreviewDTO struct {
	FileName      string `json:"fileName"`
	MIMEType      string `json:"mimeType"`
	ContentBase64 string `json:"contentBase64"`
}
type ProofDTO struct {
	ID           string  `json:"id"`
	OwnerType    string  `json:"ownerType"`
	OwnerID      string  `json:"ownerId"`
	AttachmentID string  `json:"attachmentId"`
	Status       string  `json:"status"`
	VersionLabel string  `json:"versionLabel"`
	ApproverNote string  `json:"approverNote"`
	InternalNote string  `json:"internalNote"`
	PreparedAt   *string `json:"preparedAt"`
	ApprovedAt   *string `json:"approvedAt"`
	RejectedAt   *string `json:"rejectedAt"`
	CreatedAt    *string `json:"createdAt"`
}

func (a *App) metadataService() (*application.MetadataService, error) {
	if a.startupError != nil {
		return nil, a.startupError
	}
	if a.metadata == nil {
		return nil, fmt.Errorf("metadata service is not initialized")
	}
	return a.metadata, nil
}
func (a *App) ListAttachments(ownerType, ownerID string) ([]AttachmentDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return nil, e
	}
	rows, e := s.ListAttachments(a.materialContext(), ownerType, ownerID)
	if e != nil {
		return nil, e
	}
	out := make([]AttachmentDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, AttachmentDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, FileName: r.FileName, Path: r.Path, MIMEType: r.MIMEType, SizeBytes: r.SizeBytes, Checksum: r.Checksum, Category: r.Category, Notes: r.Notes, CreatedAt: r.CreatedAt})
	}
	return out, nil
}
func (a *App) ReadAttachment(id string) (AttachmentPreviewDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return AttachmentPreviewDTO{}, e
	}
	attachment, e := s.GetAttachment(a.materialContext(), id)
	if e != nil {
		return AttachmentPreviewDTO{}, e
	}
	data, e := os.ReadFile(attachment.Path)
	if e != nil {
		return AttachmentPreviewDTO{}, fmt.Errorf("read attachment: %w", e)
	}
	mimeType := attachment.MIMEType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return AttachmentPreviewDTO{FileName: attachment.FileName, MIMEType: mimeType, ContentBase64: base64.StdEncoding.EncodeToString(data)}, nil
}
func (a *App) SaveAttachmentAs(id string) (bool, error) {
	s, e := a.metadataService()
	if e != nil {
		return false, e
	}
	if a.ctx == nil {
		return false, fmt.Errorf("application context is not initialized")
	}
	attachment, e := s.GetAttachment(a.materialContext(), id)
	if e != nil {
		return false, e
	}
	data, e := os.ReadFile(attachment.Path)
	if e != nil {
		return false, fmt.Errorf("read attachment: %w", e)
	}

	filterName := attachment.MIMEType
	if filterName == "" {
		filterName = "File"
	}
	destination, e := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                "Save attachment as",
		DefaultFilename:      attachment.FileName,
		Filters:              []runtime.FileFilter{{DisplayName: filterName, Pattern: "*"}},
		CanCreateDirectories: true,
	})
	if e != nil {
		return false, fmt.Errorf("choose attachment destination: %w", e)
	}
	if destination == "" {
		return false, nil
	}
	if e = os.WriteFile(destination, data, 0o600); e != nil {
		return false, fmt.Errorf("save attachment: %w", e)
	}
	return true, nil
}
func (a *App) AddAttachment(ownerType, ownerID, fileName, path, mimeType string, sizeBytes *int64, checksum, category, notes string) (AttachmentDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return AttachmentDTO{}, e
	}
	r, e := s.AddAttachment(a.materialContext(), ownerType, ownerID, fileName, path, mimeType, sizeBytes, checksum, category, notes)
	if e != nil {
		return AttachmentDTO{}, e
	}
	return AttachmentDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, FileName: r.FileName, Path: r.Path, MIMEType: r.MIMEType, SizeBytes: r.SizeBytes, Checksum: r.Checksum, Category: r.Category, Notes: r.Notes, CreatedAt: r.CreatedAt}, nil
}
func (a *App) ImportAttachment(ownerType, ownerID, fileName, mimeType, contentBase64, category, notes string) (AttachmentDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return AttachmentDTO{}, e
	}
	ownerType = strings.TrimSpace(ownerType)
	ownerID = strings.TrimSpace(ownerID)
	fileName = filepath.Base(strings.ReplaceAll(strings.TrimSpace(fileName), "\\", "/"))
	if ownerType != "order" || ownerID == "" || strings.ContainsAny(ownerID, `/\\`) || fileName == "" || fileName == "." {
		return AttachmentDTO{}, fmt.Errorf("invalid managed attachment details")
	}
	data, e := base64.StdEncoding.DecodeString(contentBase64)
	if e != nil {
		return AttachmentDTO{}, fmt.Errorf("decode attachment: %w", e)
	}
	if len(data) == 0 {
		return AttachmentDTO{}, fmt.Errorf("attachment file is empty")
	}

	directory := filepath.Join(a.paths.Attachments, ownerType, ownerID)
	if e = os.MkdirAll(directory, 0o700); e != nil {
		return AttachmentDTO{}, fmt.Errorf("create attachment directory: %w", e)
	}
	stored, e := os.CreateTemp(directory, "attachment-*")
	if e != nil {
		return AttachmentDTO{}, fmt.Errorf("create managed attachment: %w", e)
	}
	path := stored.Name()
	removeStored := true
	defer func() {
		if removeStored {
			_ = os.Remove(path)
		}
	}()
	if _, e = stored.Write(data); e != nil {
		_ = stored.Close()
		return AttachmentDTO{}, fmt.Errorf("store attachment: %w", e)
	}
	if e = stored.Close(); e != nil {
		return AttachmentDTO{}, fmt.Errorf("close managed attachment: %w", e)
	}
	digest := sha256.Sum256(data)
	size := int64(len(data))
	r, e := s.AddAttachment(a.materialContext(), ownerType, ownerID, fileName, path, mimeType, &size, hex.EncodeToString(digest[:]), category, notes)
	if e != nil {
		return AttachmentDTO{}, e
	}
	removeStored = false
	return AttachmentDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, FileName: r.FileName, Path: r.Path, MIMEType: r.MIMEType, SizeBytes: r.SizeBytes, Checksum: r.Checksum, Category: r.Category, Notes: r.Notes, CreatedAt: r.CreatedAt}, nil
}
func (a *App) RemoveAttachment(id string) error {
	s, e := a.metadataService()
	if e != nil {
		return e
	}
	return s.RemoveAttachment(a.materialContext(), id)
}
func (a *App) ListProofs(ownerType, ownerID string) ([]ProofDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return nil, e
	}
	rows, e := s.ListProofs(a.materialContext(), ownerType, ownerID)
	if e != nil {
		return nil, e
	}
	out := make([]ProofDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ProofDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, AttachmentID: r.AttachmentID, Status: r.Status, VersionLabel: r.VersionLabel, ApproverNote: r.ApproverNote, InternalNote: r.InternalNote, PreparedAt: r.PreparedAt, ApprovedAt: r.ApprovedAt, RejectedAt: r.RejectedAt, CreatedAt: r.CreatedAt})
	}
	return out, nil
}
func (a *App) CreateProof(ownerType, ownerID, attachmentID, status, versionLabel, approverNote, internalNote string) (ProofDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return ProofDTO{}, e
	}
	r, e := s.CreateProof(a.materialContext(), ownerType, ownerID, attachmentID, status, versionLabel, approverNote, internalNote)
	if e != nil {
		return ProofDTO{}, e
	}
	return ProofDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, AttachmentID: r.AttachmentID, Status: r.Status, VersionLabel: r.VersionLabel, ApproverNote: r.ApproverNote, InternalNote: r.InternalNote, PreparedAt: r.PreparedAt, ApprovedAt: r.ApprovedAt, RejectedAt: r.RejectedAt, CreatedAt: r.CreatedAt}, nil
}
func (a *App) UpdateProofStatus(ownerType, ownerID, proofID, status, approverNote string) (ProofDTO, error) {
	s, e := a.metadataService()
	if e != nil {
		return ProofDTO{}, e
	}
	r, e := s.TransitionProof(a.materialContext(), ownerType, ownerID, proofID, status, approverNote)
	if e != nil {
		return ProofDTO{}, e
	}
	return ProofDTO{ID: r.ID, OwnerType: r.OwnerType, OwnerID: r.OwnerID, AttachmentID: r.AttachmentID, Status: r.Status, VersionLabel: r.VersionLabel, ApproverNote: r.ApproverNote, InternalNote: r.InternalNote, PreparedAt: r.PreparedAt, ApprovedAt: r.ApprovedAt, RejectedAt: r.RejectedAt, CreatedAt: r.CreatedAt}, nil
}
