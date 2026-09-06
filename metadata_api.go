package main

import (
	"Atropaten/internal/application"
	"fmt"
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
