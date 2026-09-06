package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"Atropaten/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

func TestV5UpgradePreservesCommercialData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v5.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for version := 1; version <= 5; version++ {
		if _, err := legacy.Exec(migrations[version-1].sql); err != nil {
			t.Fatalf("migration %d: %v", version, err)
		}
		if _, err := legacy.Exec(`INSERT INTO schema_migrations(version, applied_at) VALUES (?,?)`, version, "2026-01-01T00:00:00Z"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := legacy.Exec(`INSERT INTO customers(id,name,created_at,updated_at) VALUES('CUS-v5','Legacy customer','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`INSERT INTO orders(id,order_number,customer_id,created_at,priority,commercial_status,fulfillment_status,payment_status,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at) VALUES('ORD-v5','ORD-9001','CUS-v5','2026-01-01T00:00:00Z','Normal','Confirmed','Pending','Unpaid',900,0,900,400,'2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	customer, err := store.GetCustomer(context.Background(), "CUS-v5")
	if err != nil || customer.Name != "Legacy customer" {
		t.Fatalf("customer after v6: %+v, %v", customer, err)
	}
	order, err := store.GetOrder(context.Background(), "ORD-v5")
	if err != nil || order.OrderNumber != "ORD-9001" || order.TotalRial != 900 {
		t.Fatalf("order after v6: %+v, %v", order, err)
	}
	var version int
	if err := store.db.QueryRow(`SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil || version != 9 {
		t.Fatalf("migration version=%d err=%v", version, err)
	}
}

func testQuote(now time.Time) domain.Quote {
	q := domain.NewQuote("QUOTE-1", "CUS-1", now)
	q.QuoteNumber = "QUO-TEST"
	q.Items = []domain.QuoteItem{
		{ID: "QITEM-1", QuoteID: q.ID, Position: 0, ServiceID: "SVC-old", ServiceNameSnapshot: "Saved design", ServiceCodeSnapshot: "DES", Quantity: 250001, QuantityUnit: "unit", ResolvedParametersJSON: `{"size":"A4"}`, CostBreakdownJSON: `[{"name":"Labor","amountRial":100001}]`, PricingSnapshotJSON: `{"effectiveSellingPriceRial":250001}`, EstimatedCostRial: 100001, SuggestedPriceRial: 225001, SellingPriceRial: 250001, Notes: "Keep exact"},
		{ID: "QITEM-2", QuoteID: q.ID, Position: 1, ServiceID: "SVC-new", ServiceNameSnapshot: "Saved print", ServiceCodeSnapshot: "PRN", Quantity: 500, QuantityUnit: "sheet", ResolvedParametersJSON: `{"quantity":"500"}`, CostBreakdownJSON: `[{"name":"Paper","amountRial":125000}]`, PricingSnapshotJSON: `{"effectiveSellingPriceRial":300000}`, EstimatedCostRial: 125000, SuggestedPriceRial: 300000, SellingPriceRial: 300000},
	}
	_ = q.RecalculateTotals()
	return q
}

func TestQuotePersistenceSnapshotsTotalsAndNumbering(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "quotes.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Date(2026, time.January, 12, 8, 0, 0, 0, time.UTC)
	customer, err := domain.NewCustomer("CUS-1", domain.CustomerDraft{Name: "Quote customer"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveCustomer(ctx, customer); err != nil {
		t.Fatal(err)
	}
	q := testQuote(now)
	if err := store.CreateQuote(ctx, q); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetQuote(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.QuoteNumber != "QUO-1001" || got.SubtotalRial != 550001 || got.TotalRial != 550001 || len(got.Items) != 2 || got.Items[0].Position != 0 {
		t.Fatalf("quote persistence: %+v", got)
	}
	if _, err := store.db.Exec(`UPDATE services SET name='Changed catalog' WHERE id='SVC-old'`); err != nil {
		t.Fatal(err)
	}
	reopened, err := store.GetQuote(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Items[0].ServiceNameSnapshot != "Saved design" || reopened.Items[0].PricingSnapshotJSON != q.Items[0].PricingSnapshotJSON {
		t.Fatal("quote snapshot was repriced or changed")
	}
	q2 := domain.NewQuote("QUOTE-2", "", now.Add(time.Minute))
	q2.QuoteNumber = "unused"
	if err := store.CreateQuote(ctx, q2); err != nil {
		t.Fatal(err)
	}
	got2, _ := store.GetQuote(ctx, q2.ID)
	if got2.QuoteNumber != "QUO-1002" {
		t.Fatalf("quote numbering=%s", got2.QuoteNumber)
	}
}

func TestQuoteConversionCopiesSnapshotsExactlyAndIsIdempotenceGuarded(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "convert.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	q := testQuote(now)
	if _, err := store.db.Exec(`INSERT INTO customers(id,name,created_at,updated_at) VALUES('CUS-1','Customer','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateQuote(ctx, q); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE quotes SET status='Accepted' WHERE id=?`, q.ID); err != nil {
		t.Fatal(err)
	}
	orderID, err := store.ConvertQuoteToOrder(ctx, q.ID, "ORDER-FROM-QUOTE")
	if err != nil {
		t.Fatal(err)
	}
	if orderID != "ORDER-FROM-QUOTE" {
		t.Fatal(orderID)
	}
	order, err := store.GetOrder(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	converted, err := store.GetQuote(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if converted.Status != domain.QuoteConverted || converted.ConvertedOrderID != orderID || order.QuoteID != q.ID || order.TotalRial != q.TotalRial || len(order.Items) != len(q.Items) {
		t.Fatalf("conversion mismatch: quote=%+v order=%+v", converted, order)
	}
	for i := range q.Items {
		if order.Items[i].ServiceNameSnapshot != q.Items[i].ServiceNameSnapshot || order.Items[i].ResolvedParametersJSON != q.Items[i].ResolvedParametersJSON || order.Items[i].CostBreakdownJSON != q.Items[i].CostBreakdownJSON || order.Items[i].PricingSnapshotJSON != q.Items[i].PricingSnapshotJSON || order.Items[i].SellingPriceRial != q.Items[i].SellingPriceRial {
			t.Fatalf("item %d snapshot changed", i)
		}
	}
	if _, err := store.ConvertQuoteToOrder(ctx, q.ID, "SECOND-ORDER"); err == nil {
		t.Fatal("repeated conversion unexpectedly succeeded")
	}
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM orders WHERE quote_id=?`, q.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("linked order count=%d err=%v", count, err)
	}
}

func TestQuoteConversionRollsBackWhenOrderInsertFails(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "rollback.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	q := testQuote(time.Now().UTC())
	if _, err := store.db.Exec(`INSERT INTO customers(id,name,created_at,updated_at) VALUES('CUS-1','Customer','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateQuote(ctx, q); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`UPDATE quotes SET status='Accepted' WHERE id=?`, q.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO orders(id,order_number,created_at,priority,commercial_status,fulfillment_status,payment_status,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at) VALUES('COLLISION','ORD-COLLISION','2026-01-01T00:00:00Z','Normal','Draft','Pending','Unpaid',0,0,0,0,'2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConvertQuoteToOrder(ctx, q.ID, "COLLISION"); err == nil {
		t.Fatal("collision conversion unexpectedly succeeded")
	}
	got, err := store.GetQuote(ctx, q.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.QuoteAccepted || got.ConvertedOrderID != "" {
		t.Fatalf("quote mutated after rollback: %+v", got)
	}
}

func TestAttachmentAndProofHistoryPersistence(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "metadata.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	size := int64(42)
	a := domain.Attachment{ID: "ATT-1", OwnerType: domain.AttachmentQuote, OwnerID: "QUOTE-1", FileName: "artwork.pdf", Path: "refs/artwork.pdf", MIMEType: "application/pdf", SizeBytes: &size, Checksum: "abc", Category: domain.AttachmentArtwork, CreatedAt: now}
	if err := store.SaveAttachment(ctx, a); err != nil {
		t.Fatal(err)
	}
	files, err := store.ListAttachments(ctx, domain.AttachmentQuote, "QUOTE-1")
	if err != nil || len(files) != 1 || *files[0].SizeBytes != 42 {
		t.Fatalf("attachments=%+v err=%v", files, err)
	}
	p1 := domain.Proof{ID: "PRF-1", OwnerType: domain.AttachmentQuote, OwnerID: "QUOTE-1", AttachmentID: a.ID, Status: domain.ProofReady, VersionLabel: "v1", PreparedAt: &now, CreatedAt: now}
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
	history, err := store.ListProofs(ctx, domain.AttachmentQuote, "QUOTE-1")
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
}
