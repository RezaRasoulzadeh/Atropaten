package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrProofNotFound = errors.New("proof not found")

type ProofStatus string

const (
	ProofDraft           ProofStatus = "Draft"
	ProofReady           ProofStatus = "Ready"
	ProofWaitingApproval ProofStatus = "Waiting Customer Approval"
	ProofApproved        ProofStatus = "Approved"
	ProofRejected        ProofStatus = "Rejected"
)

type Proof struct {
	ID, OwnerID, AttachmentID, VersionLabel, ApproverNote, InternalNote string
	OwnerType                                                           AttachmentOwnerType
	Status                                                              ProofStatus
	PreparedAt, ApprovedAt, RejectedAt                                  *time.Time
	CreatedAt                                                           time.Time
}

func (p Proof) Validate() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.OwnerID) == "" || strings.TrimSpace(p.VersionLabel) == "" {
		return validationError("proof", "id, owner, and version label are required")
	}
	if p.OwnerType != AttachmentQuote && p.OwnerType != AttachmentOrder {
		return validationError("ownerType", "must be quote or order")
	}
	if p.Status != ProofDraft && p.Status != ProofReady && p.Status != ProofWaitingApproval && p.Status != ProofApproved && p.Status != ProofRejected {
		return validationError("status", "is unsupported")
	}
	if p.CreatedAt.IsZero() {
		return validationError("createdAt", "is required")
	}
	if p.Status == ProofApproved && p.ApprovedAt == nil {
		return validationError("approvedAt", "is required for approved proof")
	}
	if p.Status == ProofRejected && p.RejectedAt == nil {
		return validationError("rejectedAt", "is required for rejected proof")
	}
	return nil
}
func ValidProofTransition(from, to ProofStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case ProofDraft:
		return to == ProofReady
	case ProofReady:
		return to == ProofWaitingApproval || to == ProofDraft
	case ProofWaitingApproval:
		return to == ProofApproved || to == ProofRejected
	case ProofRejected:
		return to == ProofReady
	}
	return false
}
