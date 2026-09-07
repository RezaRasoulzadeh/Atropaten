package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAttachmentNotFound  = errors.New("attachment not found")
	ErrAttachmentProtected = errors.New("attachment is referenced by proof history")
)

type AttachmentOwnerType string

const (
	AttachmentQuote AttachmentOwnerType = "quote"
	AttachmentOrder AttachmentOwnerType = "order"
)

type AttachmentCategory string

const (
	AttachmentArtwork   AttachmentCategory = "artwork"
	AttachmentProof     AttachmentCategory = "proof"
	AttachmentReference AttachmentCategory = "reference"
	AttachmentOther     AttachmentCategory = "other"
)

type Attachment struct {
	ID, OwnerID, FileName, Path, MIMEType, Checksum, Notes string
	OwnerType                                              AttachmentOwnerType
	SizeBytes                                              *int64
	Category                                               AttachmentCategory
	CreatedAt                                              time.Time
}

func (a Attachment) Validate() error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.OwnerID) == "" || strings.TrimSpace(a.FileName) == "" || strings.TrimSpace(a.Path) == "" {
		return validationError("attachment", "id, owner, file name, and path are required")
	}
	if a.OwnerType != AttachmentQuote && a.OwnerType != AttachmentOrder {
		return validationError("ownerType", "must be quote or order")
	}
	if a.Category != AttachmentArtwork && a.Category != AttachmentProof && a.Category != AttachmentReference && a.Category != AttachmentOther {
		return validationError("category", "is unsupported")
	}
	if a.SizeBytes != nil && *a.SizeBytes < 0 {
		return validationError("sizeBytes", "cannot be negative")
	}
	if a.CreatedAt.IsZero() {
		return validationError("createdAt", "is required")
	}
	return nil
}
