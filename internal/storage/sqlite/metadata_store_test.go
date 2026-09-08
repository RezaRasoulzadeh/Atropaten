package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestAttachmentAndProofHistoryPersistence(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "metadata.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	size := int64(42)
	a := domain.Attachment{ID: "ATT-1", OwnerType: domain.AttachmentOrder, OwnerID: "ORDER-1", FileName: "artwork.pdf", Path: "refs/artwork.pdf", MIMEType: "application/pdf", SizeBytes: &size, Checksum: "abc", Category: domain.AttachmentArtwork, CreatedAt: now}
	if err := store.SaveAttachment(ctx, a); err != nil {
		t.Fatal(err)
	}
	files, err := store.ListAttachments(ctx, domain.AttachmentOrder, "ORDER-1")
	if err != nil || len(files) != 1 || *files[0].SizeBytes != 42 {
		t.Fatalf("attachments=%+v err=%v", files, err)
	}
	p1 := domain.Proof{ID: "PRF-1", OwnerType: domain.AttachmentOrder, OwnerID: "ORDER-1", AttachmentID: a.ID, Status: domain.ProofReady, VersionLabel: "v1", PreparedAt: &now, CreatedAt: now}
	if err := store.SaveProof(ctx, p1); err != nil {
		t.Fatal(err)
	}
	approved := now.Add(time.Minute)
	p2 := p1
	p2.ID = "PRF-2"
	p2.Status = domain.ProofApproved
	p2.ApprovedAt = &approved
	p2.CreatedAt = approved
	if err := store.SaveProof(ctx, p2); err != nil {
		t.Fatal(err)
	}
	history, err := store.ListProofs(ctx, domain.AttachmentOrder, "ORDER-1")
	if err != nil || len(history) != 2 || history[0].Status != domain.ProofReady || history[1].Status != domain.ProofApproved {
		t.Fatalf("proof history=%+v err=%v", history, err)
	}
	invalid := p1
	invalid.ID = "PRF-invalid"
	invalid.Status = domain.ProofApproved
	invalid.ApprovedAt = nil
	if err := store.SaveProof(ctx, invalid); err == nil {
		t.Fatal("approved proof without timestamp accepted")
	}
	if err := store.DeleteAttachment(ctx, a.ID); !errors.Is(err, domain.ErrAttachmentProtected) {
		t.Fatalf("proof-linked attachment delete error=%v", err)
	}
	metadata := application.NewMetadataService(store, store)
	if _, err := metadata.CreateProof(ctx, "order", "ORDER-1", a.ID, string(domain.ProofApproved), "v2", "", ""); err == nil {
		t.Fatal("new proof version bypassed workflow")
	}
}
