package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

type AttachmentRepository interface {
	ListAttachments(context.Context, domain.AttachmentOwnerType, string) ([]domain.Attachment, error)
	SaveAttachment(context.Context, domain.Attachment) error
	DeleteAttachment(context.Context, string) error
}
type ProofRepository interface {
	ListProofs(context.Context, domain.AttachmentOwnerType, string) ([]domain.Proof, error)
	SaveProof(context.Context, domain.Proof) error
}
type AttachmentView struct {
	ID, OwnerType, OwnerID, FileName, Path, MIMEType, Checksum, Category, Notes string
	SizeBytes                                                                   *int64
	CreatedAt                                                                   string
}
type ProofView struct {
	ID, OwnerType, OwnerID, AttachmentID, Status, VersionLabel, ApproverNote, InternalNote string
	PreparedAt, ApprovedAt, RejectedAt, CreatedAt                                          *string
}
type MetadataService struct {
	attachments AttachmentRepository
	proofs      ProofRepository
	now         func() time.Time
}

func NewMetadataService(a AttachmentRepository, p ProofRepository) *MetadataService {
	return &MetadataService{attachments: a, proofs: p, now: time.Now}
}
func (s *MetadataService) ListAttachments(ctx context.Context, ownerType, ownerID string) ([]AttachmentView, error) {
	rows, e := s.attachments.ListAttachments(ctx, domain.AttachmentOwnerType(ownerType), strings.TrimSpace(ownerID))
	if e != nil {
		return nil, e
	}
	out := make([]AttachmentView, 0, len(rows))
	for _, a := range rows {
		out = append(out, attachmentView(a))
	}
	return out, nil
}
func (s *MetadataService) AddAttachment(ctx context.Context, ownerType, ownerID, fileName, path, mime string, size *int64, checksum, category, notes string) (AttachmentView, error) {
	a := domain.Attachment{ID: "", OwnerType: domain.AttachmentOwnerType(ownerType), OwnerID: strings.TrimSpace(ownerID), FileName: strings.TrimSpace(fileName), Path: strings.TrimSpace(path), MIMEType: strings.TrimSpace(mime), SizeBytes: size, Checksum: strings.TrimSpace(checksum), Category: domain.AttachmentCategory(category), Notes: strings.TrimSpace(notes), CreatedAt: s.now().UTC()}
	id, e := randomID("ATT-")
	if e != nil {
		return AttachmentView{}, e
	}
	a.ID = id
	if e = s.attachments.SaveAttachment(ctx, a); e != nil {
		return AttachmentView{}, e
	}
	return attachmentView(a), nil
}
func (s *MetadataService) RemoveAttachment(ctx context.Context, id string) error {
	return s.attachments.DeleteAttachment(ctx, strings.TrimSpace(id))
}
func (s *MetadataService) ListProofs(ctx context.Context, ownerType, ownerID string) ([]ProofView, error) {
	rows, e := s.proofs.ListProofs(ctx, domain.AttachmentOwnerType(ownerType), strings.TrimSpace(ownerID))
	if e != nil {
		return nil, e
	}
	out := make([]ProofView, 0, len(rows))
	for _, p := range rows {
		out = append(out, proofView(p))
	}
	return out, nil
}
func (s *MetadataService) CreateProof(ctx context.Context, ownerType, ownerID, attachmentID, status, version, approverNote, internalNote string) (ProofView, error) {
	now := s.now().UTC()
	requestedStatus := domain.ProofStatus(status)
	if requestedStatus != domain.ProofDraft && requestedStatus != domain.ProofReady {
		return ProofView{}, fmt.Errorf("a new proof version must start in Draft or Ready")
	}
	p := domain.Proof{OwnerType: domain.AttachmentOwnerType(ownerType), OwnerID: strings.TrimSpace(ownerID), AttachmentID: strings.TrimSpace(attachmentID), Status: domain.ProofStatus(status), VersionLabel: strings.TrimSpace(version), ApproverNote: strings.TrimSpace(approverNote), InternalNote: strings.TrimSpace(internalNote), CreatedAt: now}
	setProofTimestamp(&p, now)
	id, e := randomID("PRF-")
	if e != nil {
		return ProofView{}, e
	}
	p.ID = id
	if e = s.proofs.SaveProof(ctx, p); e != nil {
		return ProofView{}, e
	}
	return proofView(p), nil
}
func (s *MetadataService) TransitionProof(ctx context.Context, ownerType, ownerID, proofID, status, approverNote string) (ProofView, error) {
	var previous *domain.Proof
	rows, err := s.proofs.ListProofs(ctx, domain.AttachmentOwnerType(ownerType), strings.TrimSpace(ownerID))
	if err != nil {
		return ProofView{}, err
	}
	for i := range rows {
		if rows[i].ID == proofID {
			previous = &rows[i]
			break
		}
	}
	if previous == nil {
		return ProofView{}, domain.ErrProofNotFound
	}
	next := domain.ProofStatus(status)
	if !domain.ValidProofTransition(previous.Status, next) {
		return ProofView{}, fmt.Errorf("invalid proof status transition")
	}
	now := s.now().UTC()
	copy := *previous
	copy.ID, copy.Status, copy.ApproverNote, copy.CreatedAt = "", next, strings.TrimSpace(approverNote), now
	copy.PreparedAt, copy.ApprovedAt, copy.RejectedAt = nil, nil, nil
	setProofTimestamp(&copy, now)
	copy.ID, err = randomID("PRF-")
	if err != nil {
		return ProofView{}, err
	}
	if err = s.proofs.SaveProof(ctx, copy); err != nil {
		return ProofView{}, err
	}
	return proofView(copy), nil
}
func setProofTimestamp(p *domain.Proof, now time.Time) {
	switch p.Status {
	case domain.ProofReady, domain.ProofWaitingApproval, domain.ProofApproved, domain.ProofRejected:
		p.PreparedAt = &now
	}
	if p.Status == domain.ProofApproved {
		p.ApprovedAt = &now
	}
	if p.Status == domain.ProofRejected {
		p.RejectedAt = &now
	}
}
func attachmentView(a domain.Attachment) AttachmentView {
	return AttachmentView{ID: a.ID, OwnerType: string(a.OwnerType), OwnerID: a.OwnerID, FileName: a.FileName, Path: a.Path, MIMEType: a.MIMEType, SizeBytes: a.SizeBytes, Checksum: a.Checksum, Category: string(a.Category), Notes: a.Notes, CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339Nano)}
}
func proofView(p domain.Proof) ProofView {
	fmtTime := func(t *time.Time) *string {
		if t == nil {
			return nil
		}
		v := t.UTC().Format(time.RFC3339Nano)
		return &v
	}
	return ProofView{ID: p.ID, OwnerType: string(p.OwnerType), OwnerID: p.OwnerID, AttachmentID: p.AttachmentID, Status: string(p.Status), VersionLabel: p.VersionLabel, ApproverNote: p.ApproverNote, InternalNote: p.InternalNote, PreparedAt: fmtTime(p.PreparedAt), ApprovedAt: fmtTime(p.ApprovedAt), RejectedAt: fmtTime(p.RejectedAt), CreatedAt: fmtTime(&p.CreatedAt)}
}
