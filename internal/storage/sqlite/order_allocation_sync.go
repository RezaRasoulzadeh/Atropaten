package sqlite

import (
	"context"
	"database/sql"
	"time"

	"Atropaten/internal/domain"
)

// Keep the oldest allocations up to the current obligation. A different
// customer's money stays with its original owner. Cash/payment history is never
// rewritten: release allocations, append retained portions, and reclassify excess
// AR to the existing customer-credit account.
func (s *Store) reconcileOrderAllocationsTx(ctx context.Context, tx *sql.Tx, o domain.Order, replacingInvoice string) error {
	rows, err := tx.QueryContext(ctx, `SELECT a.id,a.payment_id,a.target_type,a.target_id,a.amount_rial,COALESCE(p.customer_id,'') FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.reversed=0 AND p.status='posted' AND ((a.target_type='order' AND a.target_id=?) OR (a.target_type='invoice' AND a.target_id IN (SELECT id FROM invoices WHERE order_id=?))) ORDER BY p.posted_at,p.id,a.position,a.id`, o.ID, o.ID)
	if err != nil {
		return err
	}
	type allocation struct {
		id, payment, typ, target, customer string
		amount                             int64
	}
	var allocations []allocation
	for rows.Next() {
		var a allocation
		if err = rows.Scan(&a.id, &a.payment, &a.typ, &a.target, &a.amount, &a.customer); err != nil {
			rows.Close()
			return err
		}
		allocations = append(allocations, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	remaining := o.TotalRial
	for _, a := range allocations {
		keep := min(a.amount, remaining)
		if a.customer != o.CustomerID {
			keep = 0
		}
		remaining -= keep
		move := a.typ == "invoice" && a.target == replacingInvoice
		if keep == a.amount && !move {
			continue
		}
		if _, err = tx.ExecContext(ctx, `UPDATE payment_allocations SET reversed=1 WHERE id=?`, a.id); err != nil {
			return err
		}
		if keep > 0 {
			typ, target := a.typ, a.target
			if move {
				typ, target = "order", o.ID
			}
			if _, err = tx.ExecContext(ctx, `INSERT INTO payment_allocations(id,payment_id,position,target_type,target_id,amount_rial,reversed) SELECT ?,?,COALESCE(MAX(position),-1)+1,?,?,?,0 FROM payment_allocations WHERE payment_id=?`, a.id+"-RETAINED", a.payment, typ, target, keep, a.payment); err != nil {
				return err
			}
		}
		if excess := a.amount - keep; excess > 0 {
			now := time.Now().UTC()
			id := "JE-ALLOC-CREDIT-" + a.id
			_, err = s.postJournalTx(ctx, tx, domain.JournalEntry{ID: id, Description: "Order edit: released allocation to customer credit", SourceType: "payment_allocation_adjustment", SourceID: a.payment, IdempotencyKey: id, PostedAt: now, CreatedAt: now, Lines: []domain.JournalLine{
				{ID: id + "-AR", JournalEntryID: id, Position: 0, AccountID: "ACC-AR", DebitRial: excess, PartyType: "customer", PartyID: a.customer, Memo: "Released allocation " + a.id},
				{ID: id + "-CREDIT", JournalEntryID: id, Position: 1, AccountID: "ACC-CUSTOMER-CREDIT", CreditRial: excess, PartyType: "customer", PartyID: a.customer, Memo: "Unallocated payment value"},
			}})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) reversePaymentAllocationAdjustmentsTx(ctx context.Context, tx *sql.Tx, paymentID string) error {
	ids, err := orderDependencyIDs(ctx, tx, `SELECT id FROM journal_entries WHERE source_type='payment_allocation_adjustment' AND source_id=? ORDER BY created_at,id`, paymentID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if _, err = s.reverseJournalTx(ctx, tx, id, "payment:allocation:reverse:"+id, "Reverse payment allocation adjustment", time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}
