package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"Atropaten/internal/application"
	"Atropaten/internal/domain"
)

func TestDeleteOrderCleansDependenciesAndPreservesReversedHistory(t *testing.T) {
	for _, invoiceStatus := range []string{"none", "Draft", "Posted", "Voided"} {
		t.Run(invoiceStatus, func(t *testing.T) {
			s, o, job := productionFlowFixture(t)
			ctx := context.Background()
			now := time.Now().UTC()
			var invoiceID string
			if invoiceStatus != "none" {
				inv, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
				if err != nil {
					t.Fatal(err)
				}
				invoiceID = inv.ID
				if invoiceStatus != "Draft" {
					if err = s.PostInvoice(ctx, inv.ID); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := s.TransitionProductionJob(ctx, job.ID, "In Progress"); err != nil {
				t.Fatal(err)
			}
			usage, err := s.RecordProductionConsumption(ctx, job.ID, "MAT-paper", "delete-test", 12*domain.QuantityScale, 2*domain.QuantityScale, "With waste")
			if err != nil {
				t.Fatal(err)
			}
			// Correction plus partial outsourcing must not return the same stock twice.
			if err = s.UpdateProductionConsumption(ctx, usage.ID, "delete-correction", 10*domain.QuantityScale, 2*domain.QuantityScale, "Correction"); err != nil {
				t.Fatal(err)
			}
			job.OutsourceQuantity = 5 * domain.QuantityScale
			job.OutsourceUnitCostRial = 300
			job.OutsourceFinancialAccountID = "FIN-CASH"
			if err = s.UpdateProductionOutsourcing(ctx, job); err != nil {
				t.Fatal(err)
			}
			// Completed jobs must be removable too, including auto-posted usage.
			if err = s.TransitionProductionJob(ctx, job.ID, "Completed"); err != nil {
				t.Fatal(err)
			}
			if invoiceStatus == "Voided" {
				if err = s.VoidInvoice(ctx, invoiceID); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = s.db.Exec(`INSERT INTO attachments(id,owner_type,owner_id,file_name,path,category,created_at) VALUES('ATT-delete','order',?,'art.png','/shared/art.png','artwork',?)`, o.ID, now.Format(time.RFC3339Nano)); err != nil {
				t.Fatal(err)
			}
			if _, err = s.db.Exec(`INSERT INTO proofs(id,owner_type,owner_id,attachment_id,status,version_label,created_at) VALUES('PROOF-delete','order',?,'ATT-delete','Approved','v1',?)`, o.ID, now.Format(time.RFC3339Nano)); err != nil {
				t.Fatal(err)
			}
			p := domain.Payment{ID: "PAY-delete", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 1000, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{ID: "ALLOC-delete", TargetType: "order", TargetID: o.ID, AmountRial: 1000}}}
			if _, err = s.CreatePayment(ctx, p); err != nil {
				t.Fatal(err)
			}
			before, err := s.InventoryState(ctx, "MAT-paper")
			if err != nil {
				t.Fatal(err)
			}
			if err = s.DeleteOrder(ctx, o.ID); !errors.Is(err, domain.ErrOrderDeleteProtected) {
				t.Fatalf("active payment should block: %v", err)
			}
			after, err := s.InventoryState(ctx, "MAT-paper")
			if err != nil || before != after {
				t.Fatalf("failed deletion changed stock: before=%+v after=%+v err=%v", before, after, err)
			}
			if _, err = s.ReversePayment(ctx, p.ID, ""); err != nil {
				t.Fatal(err)
			}
			if err = s.DeleteOrder(ctx, o.ID); err != nil {
				t.Fatal(err)
			}
			if _, err = s.GetOrder(ctx, o.ID); !errors.Is(err, domain.ErrOrderNotFound) {
				t.Fatalf("order still exists: %v", err)
			}
			state, err := s.InventoryState(ctx, "MAT-paper")
			if err != nil || state.PhysicalStock != 200*domain.QuantityScale || state.ReservedStock != 0 {
				t.Fatalf("stock not restored: %+v %v", state, err)
			}
			for _, table := range []string{"orders", "order_items", "production_jobs", "production_consumptions", "production_material_plans", "inventory_reservations", "attachments", "proofs"} {
				var count int
				if err = s.db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s count=%d err=%v", table, count, err)
				}
			}
			var status string
			if err = s.db.QueryRow(`SELECT status FROM expenses WHERE id=?`, "EXP-OUTSOURCE-"+job.ID).Scan(&status); err != nil || status != "Reversed" {
				t.Fatalf("expense=%s err=%v", status, err)
			}
			payment, err := s.GetPayment(ctx, p.ID)
			if err != nil || payment.Status != domain.PaymentReversedState || len(payment.Allocations) != 1 {
				t.Fatalf("payment audit lost: %+v %v", payment, err)
			}
			if invoiceStatus == "Posted" || invoiceStatus == "Voided" {
				inv, err := s.GetInvoice(ctx, invoiceID)
				if err != nil || inv.Status != "Voided" || inv.OrderID != "" || len(inv.Items) != 1 || inv.Items[0].OrderItemID != "" || inv.TotalRial != o.TotalRial {
					t.Fatalf("invoice audit=%+v err=%v", inv, err)
				}
				if _, err = s.db.Exec(`UPDATE invoices SET notes='tampered' WHERE id=?`, invoiceID); err == nil {
					t.Fatal("voided invoice is no longer immutable")
				}
				if _, err = s.db.Exec(`UPDATE invoice_items SET quantity_units=1 WHERE invoice_id=?`, invoiceID); err == nil {
					t.Fatal("voided invoice items are no longer immutable")
				}
			} else if invoiceStatus == "Draft" {
				if _, err = s.GetInvoice(ctx, invoiceID); !errors.Is(err, domain.ErrInvoiceNotFound) {
					t.Fatalf("draft retained: %v", err)
				}
			}
			var nonzero int
			if err = s.db.QueryRow(`SELECT COUNT(*) FROM (SELECT account_id FROM journal_lines GROUP BY account_id HAVING SUM(debit_rial-credit_rial)<>0)`).Scan(&nonzero); err != nil || nonzero != 0 {
				t.Fatalf("account balances not reversed: %d %v", nonzero, err)
			}
			var movements int
			if err = s.db.QueryRow(`SELECT COUNT(*) FROM inventory_movements`).Scan(&movements); err != nil {
				t.Fatal(err)
			}
			if err = s.DeleteOrder(ctx, o.ID); !errors.Is(err, domain.ErrOrderNotFound) {
				t.Fatalf("repeat deletion: %v", err)
			}
			var again int
			if err = s.db.QueryRow(`SELECT COUNT(*) FROM inventory_movements`).Scan(&again); err != nil || again != movements {
				t.Fatalf("repeat delete returned stock again: %d %d %v", movements, again, err)
			}
			rows, err := s.db.Query(`PRAGMA foreign_key_check`)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			if rows.Next() {
				t.Fatal("dangling foreign key")
			}
		})
	}
}

func TestDeleteOrderBlocksInvoicePaymentsAndRollsBackCleanupFailure(t *testing.T) {
	s, o, j := productionFlowFixture(t)
	ctx := context.Background()
	inv, err := application.NewInvoicesService(s, s).CreateFromOrder(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.PostInvoice(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	p := domain.Payment{ID: "PAY-invoice-delete", Direction: domain.PaymentIncoming, Method: domain.PaymentCash, FinancialAccountID: "FIN-CASH", AmountRial: 500, PostedAt: now, CreatedAt: now, Allocations: []domain.PaymentAllocation{{TargetType: "invoice", TargetID: inv.ID, AmountRial: 500}}}
	if _, err = s.CreatePayment(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteOrder(ctx, o.ID); !errors.Is(err, domain.ErrOrderDeleteProtected) {
		t.Fatalf("invoice payment not protected: %v", err)
	}
	if _, err = s.ReversePayment(ctx, p.ID, ""); err != nil {
		t.Fatal(err)
	}
	if err = s.TransitionProductionJob(ctx, j.ID, "In Progress"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.RecordProductionConsumption(ctx, j.ID, "MAT-paper", "rollback", 5*domain.QuantityScale, 0, ""); err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`CREATE TRIGGER fail_order_delete BEFORE DELETE ON orders BEGIN SELECT RAISE(ABORT,'test cleanup failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteOrder(ctx, o.ID); err == nil {
		t.Fatal("expected cleanup failure")
	}
	state, err := s.InventoryState(ctx, "MAT-paper")
	if err != nil || state.PhysicalStock != 195*domain.QuantityScale || state.ReservedStock != 15*domain.QuantityScale {
		t.Fatalf("partial cleanup committed: %+v %v", state, err)
	}
	invoice, err := s.GetInvoice(ctx, inv.ID)
	if err != nil || invoice.Status != "Posted" || invoice.OrderID != o.ID {
		t.Fatalf("invoice void was not rolled back: %+v %v", invoice, err)
	}
	var count int
	if err = s.db.QueryRow(`SELECT COUNT(*) FROM deleted_order_records`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("deletion audit not rolled back: %d %v", count, err)
	}
	if _, err = s.db.Exec(`DROP TRIGGER fail_order_delete`); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteOrder(ctx, o.ID); err != nil {
		t.Fatal(err)
	}
}
