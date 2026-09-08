package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Atropaten/internal/domain"
)

func (s *Store) ListAttachments(ctx context.Context, ownerType domain.AttachmentOwnerType, ownerID string) ([]domain.Attachment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,owner_type,owner_id,file_name,path,mime_type,size_bytes,checksum,category,notes,created_at FROM attachments WHERE owner_type=? AND owner_id=? ORDER BY created_at DESC,id`, string(ownerType), ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Attachment
	for rows.Next() {
		a, e := scanAttachment(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAttachment(ctx context.Context, id string) (domain.Attachment, error) {
	a, e := scanAttachment(s.db.QueryRowContext(ctx, `SELECT id,owner_type,owner_id,file_name,path,mime_type,size_bytes,checksum,category,notes,created_at FROM attachments WHERE id=?`, id))
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Attachment{}, domain.ErrAttachmentNotFound
	}
	if e != nil {
		return domain.Attachment{}, e
	}
	return a, nil
}

func (s *Store) SaveAttachment(ctx context.Context, a domain.Attachment) error {
	if err := a.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO attachments(id,owner_type,owner_id,file_name,path,mime_type,size_bytes,checksum,category,notes,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, a.ID, string(a.OwnerType), a.OwnerID, a.FileName, a.Path, a.MIMEType, a.SizeBytes, a.Checksum, string(a.Category), a.Notes, a.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save attachment: %w", err)
	}
	return nil
}

func (s *Store) DeleteAttachment(ctx context.Context, id string) error {
	var references int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM proofs WHERE attachment_id=?`, id).Scan(&references); err != nil {
		return err
	}
	if references > 0 {
		return domain.ErrAttachmentProtected
	}
	r, e := s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return domain.ErrAttachmentNotFound
	}
	return nil
}

func scanAttachment(row scanner) (domain.Attachment, error) {
	var a domain.Attachment
	var owner, category, created string
	err := row.Scan(&a.ID, &owner, &a.OwnerID, &a.FileName, &a.Path, &a.MIMEType, &a.SizeBytes, &a.Checksum, &category, &a.Notes, &created)
	a.OwnerType = domain.AttachmentOwnerType(owner)
	a.Category = domain.AttachmentCategory(category)
	if err == nil {
		a.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	}
	return a, err
}

func (s *Store) ListProofs(ctx context.Context, ownerType domain.AttachmentOwnerType, ownerID string) ([]domain.Proof, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,owner_type,owner_id,attachment_id,status,version_label,prepared_at,approved_at,rejected_at,approver_note,internal_note,created_at FROM proofs WHERE owner_type=? AND owner_id=? ORDER BY created_at ASC,id`, string(ownerType), ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Proof
	for rows.Next() {
		p, e := scanProof(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) SaveProof(ctx context.Context, p domain.Proof) error {
	if err := p.Validate(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO proofs(id,owner_type,owner_id,attachment_id,status,version_label,prepared_at,approved_at,rejected_at,approver_note,internal_note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, p.ID, string(p.OwnerType), p.OwnerID, nullableString(p.AttachmentID), string(p.Status), p.VersionLabel, nullableTime(p.PreparedAt), nullableTime(p.ApprovedAt), nullableTime(p.RejectedAt), p.ApproverNote, p.InternalNote, p.CreatedAt.UTC().Format(time.RFC3339Nano))
	return err
}

func scanProof(row scanner) (domain.Proof, error) {
	var p domain.Proof
	var owner, status, created string
	var attachment sql.NullString
	var prepared, approved, rejected sql.NullString
	err := row.Scan(&p.ID, &owner, &p.OwnerID, &attachment, &status, &p.VersionLabel, &prepared, &approved, &rejected, &p.ApproverNote, &p.InternalNote, &created)
	p.OwnerType = domain.AttachmentOwnerType(owner)
	p.AttachmentID = attachment.String
	p.Status = domain.ProofStatus(status)
	if err != nil {
		return p, err
	}
	p.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return p, err
	}
	parse := func(v string) (*time.Time, error) {
		if v == "" {
			return nil, nil
		}
		x, e := time.Parse(time.RFC3339Nano, v)
		return &x, e
	}
	if p.PreparedAt, err = parse(prepared.String); err != nil {
		return p, err
	}
	if p.ApprovedAt, err = parse(approved.String); err != nil {
		return p, err
	}
	p.RejectedAt, err = parse(rejected.String)
	return p, err
}
