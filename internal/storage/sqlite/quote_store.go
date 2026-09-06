package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"Atropaten/internal/domain"
)

func (s *Store) ListQuotes(ctx context.Context) ([]domain.Quote, error) {
	rows, err := s.db.QueryContext(ctx, quoteSelect+` ORDER BY created_at DESC, quote_number DESC`)
	if err != nil {
		return nil, fmt.Errorf("list quotes: %w", err)
	}
	defer rows.Close()
	var result []domain.Quote
	for rows.Next() {
		q, scanErr := scanQuote(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if err := s.loadQuoteItems(ctx, &q); err != nil {
			return nil, err
		}
		result = append(result, q)
	}
	return result, rows.Err()
}

const quoteSelect = `SELECT id,quote_number,customer_id,customer_name_snapshot,customer_phone_snapshot,created_at,expiry_date,status,notes,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at,converted_order_id FROM quotes`

func (s *Store) GetQuote(ctx context.Context, id string) (domain.Quote, error) {
	q, err := scanQuote(s.db.QueryRowContext(ctx, quoteSelect+` WHERE id=?`, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Quote{}, domain.ErrQuoteNotFound
	}
	if err != nil {
		return domain.Quote{}, fmt.Errorf("get quote: %w", err)
	}
	if err := s.loadQuoteItems(ctx, &q); err != nil {
		return domain.Quote{}, err
	}
	return q, nil
}

func (s *Store) CreateQuote(ctx context.Context, quote domain.Quote) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin quote create: %w", err)
	}
	rollback := func(e error) error { _ = tx.Rollback(); return e }
	if _, err := tx.ExecContext(ctx, `UPDATE quote_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return rollback(err)
	}
	var number int64
	if err := tx.QueryRowContext(ctx, `SELECT next_number-1 FROM quote_number_sequences WHERE id=1`).Scan(&number); err != nil {
		return rollback(err)
	}
	quote.QuoteNumber = fmt.Sprintf("QUO-%04d", number)
	if err := quote.Validate(); err != nil {
		return rollback(err)
	}
	if err := insertQuote(ctx, tx, quote); err != nil {
		return rollback(err)
	}
	if err := insertQuoteItems(ctx, tx, quote); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit quote create: %w", err)
	}
	return nil
}

func (s *Store) SaveQuote(ctx context.Context, quote domain.Quote) error {
	if err := quote.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(e error) error { _ = tx.Rollback(); return e }
	result, err := tx.ExecContext(ctx, `UPDATE quotes SET customer_id=?,customer_name_snapshot=?,customer_phone_snapshot=?,created_at=?,expiry_date=?,status=?,notes=?,subtotal_rial=?,discount_rial=?,total_rial=?,estimated_cost_rial=?,updated_at=?,converted_order_id=? WHERE id=?`, nullableString(quote.CustomerID), quote.CustomerNameSnapshot, quote.CustomerPhoneSnapshot, quote.CreatedAt.UTC().Format(time.RFC3339Nano), nullableTime(quote.ExpiryDate), string(quote.Status), quote.Notes, quote.SubtotalRial, quote.DiscountRial, quote.TotalRial, quote.EstimatedCostRial, quote.UpdatedAt.UTC().Format(time.RFC3339Nano), nullableString(quote.ConvertedOrderID), quote.ID)
	if err != nil {
		return rollback(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return rollback(domain.ErrQuoteNotFound)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM quote_items WHERE quote_id=?`, quote.ID); err != nil {
		return rollback(err)
	}
	if err := insertQuoteItems(ctx, tx, quote); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit quote save: %w", err)
	}
	return nil
}

func (s *Store) ConvertQuoteToOrder(ctx context.Context, quoteID, orderID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	rollback := func(e error) (string, error) { _ = tx.Rollback(); return "", e }
	q, err := scanQuote(tx.QueryRowContext(ctx, quoteSelect+` WHERE id=?`, strings.TrimSpace(quoteID)))
	if errors.Is(err, sql.ErrNoRows) {
		return rollback(domain.ErrQuoteNotFound)
	}
	if err != nil {
		return rollback(err)
	}
	if q.Status == domain.QuoteConverted || q.ConvertedOrderID != "" {
		return rollback(fmt.Errorf("quote already converted to order %s", q.ConvertedOrderID))
	}
	if q.Status != domain.QuoteAccepted {
		return rollback(fmt.Errorf("only accepted quotes can be converted"))
	}
	if err := loadQuoteItemsTx(ctx, tx, &q); err != nil {
		return rollback(err)
	}
	order := domain.NewOrder(orderID, q.CustomerID, q.CreatedAt)
	order.QuoteID, order.CustomerNameSnapshot, order.CustomerPhoneSnapshot = q.ID, q.CustomerNameSnapshot, q.CustomerPhoneSnapshot
	order.Notes, order.SubtotalRial, order.DiscountRial, order.TotalRial, order.EstimatedCostRial = q.Notes, q.SubtotalRial, q.DiscountRial, q.TotalRial, q.EstimatedCostRial
	order.CommercialStatus = domain.CommercialConfirmed
	for _, item := range q.Items {
		id, idErr := storeID("ITM-")
		if idErr != nil {
			return rollback(idErr)
		}
		order.Items = append(order.Items, domain.OrderItem{ID: id, OrderID: order.ID, Position: item.Position, ServiceID: item.ServiceID, ServiceNameSnapshot: item.ServiceNameSnapshot, ServiceCodeSnapshot: item.ServiceCodeSnapshot, Quantity: item.Quantity, QuantityUnit: item.QuantityUnit, ResolvedParametersJSON: item.ResolvedParametersJSON, CostBreakdownJSON: item.CostBreakdownJSON, PricingSnapshotJSON: item.PricingSnapshotJSON, EstimatedCostRial: item.EstimatedCostRial, SuggestedPriceRial: item.SuggestedPriceRial, SellingPriceRial: item.SellingPriceRial, Notes: item.Notes})
	}
	if _, err := tx.ExecContext(ctx, `UPDATE order_number_sequences SET next_number=next_number+1 WHERE id=1`); err != nil {
		return rollback(err)
	}
	var number int64
	if err := tx.QueryRowContext(ctx, `SELECT next_number-1 FROM order_number_sequences WHERE id=1`).Scan(&number); err != nil {
		return rollback(err)
	}
	order.OrderNumber = fmt.Sprintf("ORD-%04d", number)
	if err := order.Validate(); err != nil {
		return rollback(err)
	}
	if err := insertOrder(ctx, tx, order); err != nil {
		return rollback(err)
	}
	if err := insertOrderItems(ctx, tx, order); err != nil {
		return rollback(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE quotes SET status=?,converted_order_id=?,updated_at=? WHERE id=? AND status=? AND converted_order_id IS NULL`, string(domain.QuoteConverted), order.ID, time.Now().UTC().Format(time.RFC3339Nano), q.ID, string(domain.QuoteAccepted)); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit quote conversion: %w", err)
	}
	return order.ID, nil
}

func insertQuote(ctx context.Context, tx *sql.Tx, q domain.Quote) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO quotes(id,quote_number,customer_id,customer_name_snapshot,customer_phone_snapshot,created_at,expiry_date,status,notes,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at,converted_order_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, q.ID, q.QuoteNumber, nullableString(q.CustomerID), q.CustomerNameSnapshot, q.CustomerPhoneSnapshot, q.CreatedAt.UTC().Format(time.RFC3339Nano), nullableTime(q.ExpiryDate), string(q.Status), q.Notes, q.SubtotalRial, q.DiscountRial, q.TotalRial, q.EstimatedCostRial, q.UpdatedAt.UTC().Format(time.RFC3339Nano), nullableString(q.ConvertedOrderID))
	if err != nil {
		return fmt.Errorf("insert quote: %w", err)
	}
	return nil
}
func insertQuoteItems(ctx context.Context, tx *sql.Tx, q domain.Quote) error {
	for _, item := range q.Items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO quote_items(id,quote_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID, q.ID, item.Position, item.ServiceID, item.ServiceNameSnapshot, item.ServiceCodeSnapshot, int64(item.Quantity), item.QuantityUnit, item.ResolvedParametersJSON, item.CostBreakdownJSON, item.PricingSnapshotJSON, item.EstimatedCostRial, item.SuggestedPriceRial, item.SellingPriceRial, item.Notes); err != nil {
			return fmt.Errorf("insert quote item: %w", err)
		}
	}
	return nil
}
func (s *Store) loadQuoteItems(ctx context.Context, q *domain.Quote) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := loadQuoteItemsTx(ctx, tx, q); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func loadQuoteItemsTx(ctx context.Context, tx *sql.Tx, q *domain.Quote) error {
	rows, err := tx.QueryContext(ctx, `SELECT id,quote_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes FROM quote_items WHERE quote_id=? ORDER BY display_order,id`, q.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		item, scanErr := scanQuoteItem(rows)
		if scanErr != nil {
			return scanErr
		}
		q.Items = append(q.Items, item)
	}
	return rows.Err()
}

func scanQuote(row scanner) (domain.Quote, error) {
	var q domain.Quote
	var customerID, convertedID, expiry sql.NullString
	var created, updated, status string
	if err := row.Scan(&q.ID, &q.QuoteNumber, &customerID, &q.CustomerNameSnapshot, &q.CustomerPhoneSnapshot, &created, &expiry, &status, &q.Notes, &q.SubtotalRial, &q.DiscountRial, &q.TotalRial, &q.EstimatedCostRial, &updated, &convertedID); err != nil {
		return q, err
	}
	q.CustomerID = customerID.String
	q.ConvertedOrderID = convertedID.String
	q.Status = domain.QuoteStatus(status)
	var err error
	q.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return q, err
	}
	q.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return q, err
	}
	if expiry.Valid {
		x, e := time.Parse(time.RFC3339Nano, expiry.String)
		if e != nil {
			return q, e
		}
		q.ExpiryDate = &x
	}
	return q, nil
}
func scanQuoteItem(row scanner) (domain.QuoteItem, error) {
	var item domain.QuoteItem
	var quantity int64
	err := row.Scan(&item.ID, &item.QuoteID, &item.Position, &item.ServiceID, &item.ServiceNameSnapshot, &item.ServiceCodeSnapshot, &quantity, &item.QuantityUnit, &item.ResolvedParametersJSON, &item.CostBreakdownJSON, &item.PricingSnapshotJSON, &item.EstimatedCostRial, &item.SuggestedPriceRial, &item.SellingPriceRial, &item.Notes)
	item.Quantity = domain.Quantity(quantity)
	return item, err
}

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
func storeID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b), nil
}
