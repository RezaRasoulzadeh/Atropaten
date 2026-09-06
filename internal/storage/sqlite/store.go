package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Atropaten/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}
	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	if err := store.ensureLegacyOpeningMovements(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	if err := store.seedAccounting(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create migration metadata: %w", err)
	}
	var current int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current); err != nil {
		return fmt.Errorf("read migration version: %w", err)
	}
	for _, migration := range migrations {
		if migration.version <= current {
			continue
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES (?, ?)`, migration.version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", migration.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", migration.version, err)
		}
		current = migration.version
	}
	return nil
}

type migration struct {
	version int
	sql     string
}

var migrations = []migration{{
	version: 1,
	sql: `CREATE TABLE materials (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL CHECK(length(trim(name)) > 0),
		sku TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT '',
		purchase_unit TEXT NOT NULL,
		consumption_unit TEXT NOT NULL,
		conversion_factor_units INTEGER NOT NULL CHECK(conversion_factor_units > 0),
		physical_stock_units INTEGER NOT NULL CHECK(physical_stock_units >= 0),
		reorder_level_units INTEGER NOT NULL CHECK(reorder_level_units >= 0),
		average_unit_cost_rial INTEGER NOT NULL CHECK(average_unit_cost_rial >= 0),
		preferred_supplier TEXT NOT NULL DEFAULT '',
		notes TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)),
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`}, {
	version: 2,
	sql: `CREATE TABLE services (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL CHECK(length(trim(name)) > 0),
		code TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)),
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE service_parameters (
		id TEXT PRIMARY KEY,
		service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
		parameter_key TEXT NOT NULL,
		label TEXT NOT NULL CHECK(length(trim(label)) > 0),
		parameter_type TEXT NOT NULL CHECK(parameter_type IN ('integer', 'decimal', 'boolean', 'choice', 'material-reference')),
		required INTEGER NOT NULL DEFAULT 0 CHECK(required IN (0, 1)),
		display_order INTEGER NOT NULL CHECK(display_order >= 0),
		default_value TEXT NOT NULL DEFAULT '',
		min_value_units INTEGER,
		max_value_units INTEGER,
		unit_label TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)),
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		UNIQUE(service_id, parameter_key)
	);
	CREATE TABLE service_parameter_options (
		parameter_id TEXT NOT NULL REFERENCES service_parameters(id) ON DELETE CASCADE,
		option_order INTEGER NOT NULL CHECK(option_order >= 0),
		value TEXT NOT NULL CHECK(length(trim(value)) > 0),
		PRIMARY KEY(parameter_id, option_order)
	)`}, {
	version: 3,
	sql: `CREATE TABLE machines (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL CHECK(length(trim(name)) > 0),
		code TEXT NOT NULL DEFAULT '',
		category TEXT NOT NULL DEFAULT '',
		rate_basis TEXT NOT NULL CHECK(rate_basis IN ('unit', 'minute', 'hour')),
		rate_rial INTEGER NOT NULL CHECK(rate_rial >= 0),
		setup_cost_rial INTEGER NOT NULL DEFAULT 0 CHECK(setup_cost_rial >= 0),
		notes TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)),
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE service_cost_components (
		id TEXT PRIMARY KEY,
		service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
		component_name TEXT NOT NULL CHECK(length(trim(component_name)) > 0),
		component_type TEXT NOT NULL CHECK(component_type IN ('material', 'machine', 'labor', 'outsourced', 'fixed', 'overhead', 'waste', 'manual')),
		reference_id TEXT NOT NULL DEFAULT '',
		usage_mode TEXT NOT NULL CHECK(usage_mode IN ('fixed', 'parameter')),
		parameter_key TEXT NOT NULL DEFAULT '',
		multiplier_units INTEGER NOT NULL CHECK(multiplier_units > 0),
		rate_rial INTEGER NOT NULL DEFAULT 0 CHECK(rate_rial >= 0),
		percentage_units INTEGER NOT NULL DEFAULT 0 CHECK(percentage_units >= 0),
		rate_basis TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0, 1)),
		display_order INTEGER NOT NULL CHECK(display_order >= 0),
		notes TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`}, {
	version: 4,
	sql: `ALTER TABLE service_cost_components ADD COLUMN usage_quantity_units INTEGER NOT NULL DEFAULT 1000000;
	CREATE TABLE service_pricing_rules (
		id TEXT PRIMARY KEY,
		service_id TEXT NOT NULL UNIQUE REFERENCES services(id) ON DELETE CASCADE,
		rule_type TEXT NOT NULL CHECK(rule_type IN ('fixed', 'markup', 'fixed-margin', 'per-unit', 'quantity-tiers', 'manual')),
		fixed_price_rial INTEGER NOT NULL DEFAULT 0 CHECK(fixed_price_rial >= 0),
		markup_percentage_units INTEGER NOT NULL DEFAULT 0 CHECK(markup_percentage_units >= 0),
		fixed_margin_rial INTEGER NOT NULL DEFAULT 0 CHECK(fixed_margin_rial >= 0),
		per_unit_rate_rial INTEGER NOT NULL DEFAULT 0 CHECK(per_unit_rate_rial >= 0),
		parameter_key TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	CREATE TABLE service_pricing_tiers (
		rule_id TEXT NOT NULL REFERENCES service_pricing_rules(id) ON DELETE CASCADE,
		display_order INTEGER NOT NULL CHECK(display_order >= 0),
		minimum_quantity_units INTEGER NOT NULL CHECK(minimum_quantity_units >= 0),
		price_rial INTEGER NOT NULL CHECK(price_rial >= 0),
		PRIMARY KEY(rule_id, display_order)
	)`}, {
	version: 5,
	sql: `CREATE TABLE customers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL CHECK(length(trim(name)) > 0),
		phone TEXT NOT NULL DEFAULT '', email TEXT NOT NULL DEFAULT '', address TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0, 1)), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE order_number_sequences (id INTEGER PRIMARY KEY CHECK(id = 1), next_number INTEGER NOT NULL CHECK(next_number > 0));
	INSERT INTO order_number_sequences(id, next_number) VALUES (1, 1001);
	CREATE TABLE orders (
		id TEXT PRIMARY KEY, order_number TEXT NOT NULL UNIQUE, customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL,
		customer_name_snapshot TEXT NOT NULL DEFAULT '', customer_phone_snapshot TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, promised_at TEXT,
		priority TEXT NOT NULL CHECK(priority IN ('Urgent','High','Normal','Low')),
		commercial_status TEXT NOT NULL CHECK(commercial_status IN ('Draft','Confirmed','Closed','Cancelled')),
		fulfillment_status TEXT NOT NULL CHECK(fulfillment_status IN ('Pending','In Production','Ready','Delivered')),
		payment_status TEXT NOT NULL CHECK(payment_status IN ('Unpaid','Partially Paid','Paid')),
		notes TEXT NOT NULL DEFAULT '', subtotal_rial INTEGER NOT NULL CHECK(subtotal_rial >= 0), discount_rial INTEGER NOT NULL CHECK(discount_rial >= 0),
		total_rial INTEGER NOT NULL CHECK(total_rial >= 0), estimated_cost_rial INTEGER NOT NULL CHECK(estimated_cost_rial >= 0), updated_at TEXT NOT NULL
	);
	CREATE TABLE order_items (
		id TEXT PRIMARY KEY, order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE CASCADE, display_order INTEGER NOT NULL CHECK(display_order >= 0),
		service_id TEXT NOT NULL DEFAULT '', service_name_snapshot TEXT NOT NULL, service_code_snapshot TEXT NOT NULL DEFAULT '', quantity_units INTEGER NOT NULL CHECK(quantity_units >= 0), quantity_unit TEXT NOT NULL DEFAULT 'unit',
		resolved_parameters_json TEXT NOT NULL, cost_breakdown_json TEXT NOT NULL, pricing_snapshot_json TEXT NOT NULL,
		estimated_cost_rial INTEGER NOT NULL CHECK(estimated_cost_rial >= 0), suggested_price_rial INTEGER NOT NULL CHECK(suggested_price_rial >= 0), selling_price_rial INTEGER NOT NULL CHECK(selling_price_rial >= 0), notes TEXT NOT NULL DEFAULT '',
		UNIQUE(order_id, display_order)
	)`}, {
	version: 6,
	sql: `CREATE TABLE quote_number_sequences (id INTEGER PRIMARY KEY CHECK(id = 1), next_number INTEGER NOT NULL CHECK(next_number > 0));
	INSERT INTO quote_number_sequences(id, next_number) VALUES (1, 1001);
	CREATE TABLE quotes (
		id TEXT PRIMARY KEY, quote_number TEXT NOT NULL UNIQUE, customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL,
		customer_name_snapshot TEXT NOT NULL DEFAULT '', customer_phone_snapshot TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, expiry_date TEXT,
		status TEXT NOT NULL CHECK(status IN ('Draft','Sent','Accepted','Rejected','Expired','Converted')), notes TEXT NOT NULL DEFAULT '',
		subtotal_rial INTEGER NOT NULL CHECK(subtotal_rial >= 0), discount_rial INTEGER NOT NULL CHECK(discount_rial >= 0), total_rial INTEGER NOT NULL CHECK(total_rial >= 0), estimated_cost_rial INTEGER NOT NULL CHECK(estimated_cost_rial >= 0), updated_at TEXT NOT NULL, converted_order_id TEXT UNIQUE
	);
	CREATE TABLE quote_items (
		id TEXT PRIMARY KEY, quote_id TEXT NOT NULL REFERENCES quotes(id) ON DELETE CASCADE, display_order INTEGER NOT NULL CHECK(display_order >= 0), service_id TEXT NOT NULL DEFAULT '', service_name_snapshot TEXT NOT NULL, service_code_snapshot TEXT NOT NULL DEFAULT '', quantity_units INTEGER NOT NULL CHECK(quantity_units >= 0), quantity_unit TEXT NOT NULL DEFAULT 'unit', resolved_parameters_json TEXT NOT NULL, cost_breakdown_json TEXT NOT NULL, pricing_snapshot_json TEXT NOT NULL, estimated_cost_rial INTEGER NOT NULL CHECK(estimated_cost_rial >= 0), suggested_price_rial INTEGER NOT NULL CHECK(suggested_price_rial >= 0), selling_price_rial INTEGER NOT NULL CHECK(selling_price_rial >= 0), notes TEXT NOT NULL DEFAULT '', UNIQUE(quote_id, display_order)
	);
	ALTER TABLE orders ADD COLUMN quote_id TEXT REFERENCES quotes(id) ON DELETE SET NULL;
	CREATE UNIQUE INDEX orders_quote_id_unique ON orders(quote_id) WHERE quote_id IS NOT NULL;
	CREATE INDEX quotes_created_at ON quotes(created_at DESC, quote_number DESC);
	CREATE TABLE attachments (id TEXT PRIMARY KEY, owner_type TEXT NOT NULL CHECK(owner_type IN ('quote','order')), owner_id TEXT NOT NULL, file_name TEXT NOT NULL CHECK(length(trim(file_name)) > 0), path TEXT NOT NULL CHECK(length(trim(path)) > 0), mime_type TEXT NOT NULL DEFAULT '', size_bytes INTEGER CHECK(size_bytes IS NULL OR size_bytes >= 0), checksum TEXT NOT NULL DEFAULT '', category TEXT NOT NULL CHECK(category IN ('artwork','proof','reference','other')), notes TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
	CREATE INDEX attachments_owner ON attachments(owner_type, owner_id, created_at DESC);
	CREATE TABLE proofs (id TEXT PRIMARY KEY, owner_type TEXT NOT NULL CHECK(owner_type IN ('quote','order')), owner_id TEXT NOT NULL, attachment_id TEXT REFERENCES attachments(id) ON DELETE SET NULL, status TEXT NOT NULL CHECK(status IN ('Draft','Ready','Waiting Customer Approval','Approved','Rejected')), version_label TEXT NOT NULL, prepared_at TEXT, approved_at TEXT, rejected_at TEXT, approver_note TEXT NOT NULL DEFAULT '', internal_note TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
		CREATE INDEX proofs_owner ON proofs(owner_type, owner_id, created_at DESC);`,
},
	{
		version: 7,
		sql: `CREATE TABLE suppliers (
		id TEXT PRIMARY KEY, name TEXT NOT NULL CHECK(length(trim(name)) > 0), code TEXT NOT NULL DEFAULT '',
		phone TEXT NOT NULL DEFAULT '', email TEXT NOT NULL DEFAULT '', address TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0,1)), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE INDEX suppliers_name ON suppliers(lower(name), id);
	CREATE TABLE purchase_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
	INSERT INTO purchase_number_sequences(id, next_number) VALUES (1, 1001);
	CREATE TABLE purchases (
		id TEXT PRIMARY KEY, purchase_number TEXT NOT NULL UNIQUE, supplier_id TEXT NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
		supplier_name_snapshot TEXT NOT NULL DEFAULT '', supplier_code_snapshot TEXT NOT NULL DEFAULT '', supplier_invoice_number TEXT NOT NULL DEFAULT '',
		purchase_date TEXT NOT NULL, status TEXT NOT NULL CHECK(status IN ('Draft','Posted','Cancelled')), notes TEXT NOT NULL DEFAULT '',
		subtotal_rial INTEGER NOT NULL CHECK(subtotal_rial >= 0), discount_rial INTEGER NOT NULL CHECK(discount_rial >= 0),
		shipping_rial INTEGER NOT NULL CHECK(shipping_rial >= 0), tax_rial INTEGER NOT NULL CHECK(tax_rial >= 0), additional_costs_rial INTEGER NOT NULL CHECK(additional_costs_rial >= 0),
		total_rial INTEGER NOT NULL CHECK(total_rial >= 0), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE INDEX purchases_date ON purchases(purchase_date DESC, purchase_number DESC);
	CREATE TABLE purchase_items (
		id TEXT PRIMARY KEY, purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE, position INTEGER NOT NULL CHECK(position >= 0),
		material_id TEXT NOT NULL REFERENCES materials(id) ON DELETE RESTRICT, material_name_snapshot TEXT NOT NULL, purchase_unit_snapshot TEXT NOT NULL,
		consumption_unit_snapshot TEXT NOT NULL, purchase_quantity_units INTEGER NOT NULL CHECK(purchase_quantity_units >= 0), conversion_factor_units INTEGER NOT NULL CHECK(conversion_factor_units > 0),
		consumption_quantity_units INTEGER NOT NULL CHECK(consumption_quantity_units >= 0), unit_acquisition_cost_rial INTEGER NOT NULL CHECK(unit_acquisition_cost_rial >= 0),
		allocated_additional_cost_rial INTEGER NOT NULL, landed_unit_cost_rial INTEGER NOT NULL CHECK(landed_unit_cost_rial >= 0),
		line_total_rial INTEGER NOT NULL CHECK(line_total_rial >= 0), notes TEXT NOT NULL DEFAULT '', UNIQUE(purchase_id, position)
	);
	CREATE TABLE inventory_movements (
		id TEXT PRIMARY KEY, material_id TEXT NOT NULL REFERENCES materials(id) ON DELETE RESTRICT, occurred_at TEXT NOT NULL,
		movement_type TEXT NOT NULL CHECK(movement_type IN ('opening_balance','purchase','adjustment','supplier_return','production_consumption','waste','customer_return','transfer')),
		quantity_delta_units INTEGER NOT NULL, unit_cost_rial INTEGER NOT NULL CHECK(unit_cost_rial >= 0), total_cost_rial INTEGER NOT NULL,
		reference_type TEXT NOT NULL DEFAULT '', reference_id TEXT NOT NULL DEFAULT '', note TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL
	);
	CREATE INDEX inventory_movements_material_date ON inventory_movements(material_id, occurred_at, id);
	CREATE INDEX purchase_items_material ON purchase_items(material_id);
	CREATE TRIGGER inventory_movements_immutable_update BEFORE UPDATE ON inventory_movements BEGIN SELECT RAISE(ABORT, 'inventory movements are immutable'); END;
	CREATE TRIGGER inventory_movements_immutable_delete BEFORE DELETE ON inventory_movements BEGIN SELECT RAISE(ABORT, 'inventory movements are immutable'); END;`,
	},
	{
		version: 8,
		sql: `CREATE TABLE inventory_reservations (
			id TEXT PRIMARY KEY, material_id TEXT NOT NULL REFERENCES materials(id) ON DELETE RESTRICT,
			order_id TEXT REFERENCES orders(id) ON DELETE RESTRICT, order_item_id TEXT REFERENCES order_items(id) ON DELETE RESTRICT,
			production_job_id TEXT, quantity_units INTEGER NOT NULL CHECK(quantity_units > 0),
			status TEXT NOT NULL CHECK(status IN ('active','released','consumed','cancelled')),
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE INDEX inventory_reservations_material_status ON inventory_reservations(material_id,status);
		CREATE INDEX inventory_reservations_job ON inventory_reservations(production_job_id,status);
		CREATE TABLE production_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO production_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE production_jobs (
			id TEXT PRIMARY KEY, job_number TEXT NOT NULL UNIQUE, order_id TEXT NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
			order_item_id TEXT NOT NULL REFERENCES order_items(id) ON DELETE RESTRICT, service_name_snapshot TEXT NOT NULL DEFAULT '',
			quantity_units INTEGER NOT NULL CHECK(quantity_units > 0), quantity_unit TEXT NOT NULL DEFAULT 'unit', assigned_machine_id TEXT REFERENCES machines(id) ON DELETE SET NULL,
			status TEXT NOT NULL CHECK(status IN ('Pending','Ready','In Progress','Paused','Completed','Cancelled','Failed')), priority TEXT NOT NULL DEFAULT 'Normal',
			planned_at TEXT, started_at TEXT, completed_at TEXT, notes TEXT NOT NULL DEFAULT '', estimated_cost_rial INTEGER NOT NULL CHECK(estimated_cost_rial >= 0),
			actual_material_cost_rial INTEGER NOT NULL DEFAULT 0 CHECK(actual_material_cost_rial >= 0), actual_waste_cost_rial INTEGER NOT NULL DEFAULT 0 CHECK(actual_waste_cost_rial >= 0), actual_outsourced_cost_rial INTEGER NOT NULL DEFAULT 0 CHECK(actual_outsourced_cost_rial >= 0),
			outsource_supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL, outsource_description TEXT NOT NULL DEFAULT '', outsource_quoted_cost_rial INTEGER NOT NULL DEFAULT 0 CHECK(outsource_quoted_cost_rial >= 0), outsource_sent_at TEXT, outsource_expected_return_at TEXT, outsource_received_at TEXT, outsource_notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE INDEX production_jobs_queue ON production_jobs(status,priority,planned_at);
		CREATE TABLE production_consumptions (
			id TEXT PRIMARY KEY, production_job_id TEXT NOT NULL REFERENCES production_jobs(id) ON DELETE RESTRICT, material_id TEXT NOT NULL REFERENCES materials(id) ON DELETE RESTRICT,
			idempotency_key TEXT NOT NULL, consumed_quantity_units INTEGER NOT NULL CHECK(consumed_quantity_units >= 0), waste_quantity_units INTEGER NOT NULL CHECK(waste_quantity_units >= 0),
			unit_cost_rial INTEGER NOT NULL CHECK(unit_cost_rial >= 0), material_cost_rial INTEGER NOT NULL CHECK(material_cost_rial >= 0), waste_cost_rial INTEGER NOT NULL CHECK(waste_cost_rial >= 0), notes TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL,
			UNIQUE(production_job_id,idempotency_key)
		);
		CREATE INDEX production_consumptions_job ON production_consumptions(production_job_id,created_at);`,
	},
	{
		version: 9,
		sql: `CREATE TABLE accounts (
			id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('asset','liability','equity','revenue','expense')),
			parent_id TEXT REFERENCES accounts(id) ON DELETE RESTRICT, active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0,1)),
			system INTEGER NOT NULL DEFAULT 0 CHECK(system IN (0,1)), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE journal_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO journal_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE journal_entries (
			id TEXT PRIMARY KEY, entry_number TEXT NOT NULL UNIQUE, posted_at TEXT NOT NULL, description TEXT NOT NULL,
			source_type TEXT NOT NULL DEFAULT '', source_id TEXT NOT NULL DEFAULT '', idempotency_key TEXT NOT NULL UNIQUE,
			reversal_of_id TEXT REFERENCES journal_entries(id) ON DELETE RESTRICT, created_at TEXT NOT NULL
		);
		CREATE TABLE journal_lines (
			id TEXT PRIMARY KEY, journal_entry_id TEXT NOT NULL REFERENCES journal_entries(id) ON DELETE RESTRICT,
			position INTEGER NOT NULL CHECK(position >= 0), account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
			debit_rial INTEGER NOT NULL DEFAULT 0 CHECK(debit_rial >= 0), credit_rial INTEGER NOT NULL DEFAULT 0 CHECK(credit_rial >= 0),
			party_type TEXT NOT NULL DEFAULT '', party_id TEXT NOT NULL DEFAULT '', memo TEXT NOT NULL DEFAULT '',
			UNIQUE(journal_entry_id, position), CHECK((debit_rial > 0 AND credit_rial = 0) OR (credit_rial > 0 AND debit_rial = 0))
		);
		CREATE INDEX journal_entries_posted_at ON journal_entries(posted_at DESC, entry_number DESC);
		CREATE INDEX journal_lines_account ON journal_lines(account_id, journal_entry_id);
		CREATE TRIGGER journal_entries_immutable_update BEFORE UPDATE ON journal_entries BEGIN SELECT RAISE(ABORT, 'journal entries are immutable'); END;
		CREATE TRIGGER journal_entries_immutable_delete BEFORE DELETE ON journal_entries BEGIN SELECT RAISE(ABORT, 'journal entries are immutable'); END;
		CREATE TRIGGER journal_lines_immutable_update BEFORE UPDATE ON journal_lines BEGIN SELECT RAISE(ABORT, 'journal lines are immutable'); END;
		CREATE TRIGGER journal_lines_immutable_delete BEFORE DELETE ON journal_lines BEGIN SELECT RAISE(ABORT, 'journal lines are immutable'); END;
		CREATE TABLE financial_accounts (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, type TEXT NOT NULL CHECK(type IN ('cash','bank')),
			ledger_account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT, details TEXT NOT NULL DEFAULT '', active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0,1)),
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TABLE payments (
			id TEXT PRIMARY KEY, payment_number TEXT NOT NULL UNIQUE, direction TEXT NOT NULL CHECK(direction IN ('incoming','outgoing')),
			method TEXT NOT NULL, amount_rial INTEGER NOT NULL CHECK(amount_rial > 0), posted_at TEXT NOT NULL,
			financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
			customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL, supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL,
			reference TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL CHECK(status IN ('posted','reversed')),
			journal_entry_id TEXT NOT NULL UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT, idempotency_key TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL
		);
		CREATE TABLE payment_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO payment_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE payment_allocations (
			id TEXT PRIMARY KEY, payment_id TEXT NOT NULL REFERENCES payments(id) ON DELETE RESTRICT, position INTEGER NOT NULL CHECK(position >= 0),
			target_type TEXT NOT NULL CHECK(target_type IN ('order','purchase')), target_id TEXT NOT NULL, amount_rial INTEGER NOT NULL CHECK(amount_rial > 0),
			reversed INTEGER NOT NULL DEFAULT 0 CHECK(reversed IN (0,1)), UNIQUE(payment_id, position)
		);
		CREATE INDEX payments_posted_at ON payments(posted_at DESC, payment_number DESC);
		CREATE INDEX payment_allocations_target ON payment_allocations(target_type, target_id, reversed);
		ALTER TABLE purchases ADD COLUMN accounting_journal_entry_id TEXT REFERENCES journal_entries(id) ON DELETE RESTRICT;
		CREATE UNIQUE INDEX purchases_accounting_journal_unique ON purchases(accounting_journal_entry_id) WHERE accounting_journal_entry_id IS NOT NULL;`,
	},
	{
		version: 10,
		sql: `CREATE TABLE invoice_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO invoice_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE invoices (
			id TEXT PRIMARY KEY, invoice_number TEXT NOT NULL UNIQUE, customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL,
			customer_name_snapshot TEXT NOT NULL, customer_phone_snapshot TEXT NOT NULL DEFAULT '', order_id TEXT REFERENCES orders(id) ON DELETE RESTRICT,
			issue_date TEXT NOT NULL, due_date TEXT, status TEXT NOT NULL CHECK(status IN ('Draft','Posted','Partially Paid','Paid','Voided')),
			notes TEXT NOT NULL DEFAULT '', subtotal_rial INTEGER NOT NULL CHECK(subtotal_rial >= 0), discount_rial INTEGER NOT NULL CHECK(discount_rial >= 0), total_rial INTEGER NOT NULL CHECK(total_rial >= 0),
			accounting_journal_entry_id TEXT UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT, cogs_journal_entry_id TEXT UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL, CHECK(total_rial = subtotal_rial - discount_rial)
		);
		CREATE UNIQUE INDEX invoices_order_unique ON invoices(order_id) WHERE order_id IS NOT NULL;
		CREATE INDEX invoices_issue_date ON invoices(issue_date DESC,invoice_number DESC);
		CREATE TABLE invoice_items (
			id TEXT PRIMARY KEY, invoice_id TEXT NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT, position INTEGER NOT NULL CHECK(position >= 0),
			order_item_id TEXT REFERENCES order_items(id) ON DELETE RESTRICT, description_snapshot TEXT NOT NULL, service_id TEXT NOT NULL DEFAULT '',
			quantity_units INTEGER NOT NULL CHECK(quantity_units >= 0), quantity_unit TEXT NOT NULL DEFAULT 'unit', unit_price_rial INTEGER NOT NULL CHECK(unit_price_rial >= 0),
			line_total_rial INTEGER NOT NULL CHECK(line_total_rial >= 0), notes TEXT NOT NULL DEFAULT '', UNIQUE(invoice_id,position)
		);
		CREATE TRIGGER invoices_immutable_delete BEFORE DELETE ON invoices WHEN OLD.status <> 'Draft' BEGIN SELECT RAISE(ABORT,'posted invoices cannot be deleted'); END;
		CREATE TRIGGER invoices_immutable_update BEFORE UPDATE ON invoices WHEN OLD.status NOT IN ('Draft','Posted') OR (OLD.status='Posted' AND NEW.status<>'Voided') BEGIN SELECT RAISE(ABORT,'posted invoices are immutable; use void'); END;
		CREATE TRIGGER invoice_items_immutable_update BEFORE UPDATE ON invoice_items WHEN EXISTS(SELECT 1 FROM invoices WHERE id=OLD.invoice_id AND status<>'Draft') BEGIN SELECT RAISE(ABORT,'posted invoice lines are immutable'); END;
		CREATE TRIGGER invoice_items_immutable_delete BEFORE DELETE ON invoice_items WHEN EXISTS(SELECT 1 FROM invoices WHERE id=OLD.invoice_id AND status<>'Draft') BEGIN SELECT RAISE(ABORT,'posted invoice lines are immutable'); END;
		DROP INDEX payment_allocations_target;
		ALTER TABLE payment_allocations RENAME TO payment_allocations_v9;
		CREATE TABLE payment_allocations (
			id TEXT PRIMARY KEY, payment_id TEXT NOT NULL REFERENCES payments(id) ON DELETE RESTRICT, position INTEGER NOT NULL CHECK(position >= 0),
			target_type TEXT NOT NULL CHECK(target_type IN ('order','purchase','invoice')), target_id TEXT NOT NULL, amount_rial INTEGER NOT NULL CHECK(amount_rial > 0),
			reversed INTEGER NOT NULL DEFAULT 0 CHECK(reversed IN (0,1)), UNIQUE(payment_id,position)
		);
		INSERT INTO payment_allocations(id,payment_id,position,target_type,target_id,amount_rial,reversed) SELECT id,payment_id,position,target_type,target_id,amount_rial,reversed FROM payment_allocations_v9;
		DROP TABLE payment_allocations_v9;
		CREATE INDEX payment_allocations_target ON payment_allocations(target_type,target_id,reversed);
		CREATE TABLE expense_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO expense_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE expenses (
			id TEXT PRIMARY KEY, expense_number TEXT NOT NULL UNIQUE, expense_date TEXT NOT NULL, category_account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
			payee TEXT NOT NULL DEFAULT '', supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL, description TEXT NOT NULL, amount_rial INTEGER NOT NULL CHECK(amount_rial > 0),
			payment_method TEXT NOT NULL, financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT, notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL CHECK(status IN ('Posted','Reversed')),
			journal_entry_id TEXT NOT NULL UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT, idempotency_key TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE TRIGGER expenses_immutable_delete BEFORE DELETE ON expenses BEGIN SELECT RAISE(ABORT,'posted expenses cannot be deleted'); END;
		CREATE TABLE transfer_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL CHECK(next_number > 0));
		INSERT INTO transfer_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE financial_transfers (
			id TEXT PRIMARY KEY, transfer_number TEXT NOT NULL UNIQUE, source_financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
			destination_financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT, amount_rial INTEGER NOT NULL CHECK(amount_rial > 0), transfer_date TEXT NOT NULL,
			reference TEXT NOT NULL DEFAULT '', notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL CHECK(status IN ('Posted','Reversed')), journal_entry_id TEXT NOT NULL UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT,
			idempotency_key TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, CHECK(source_financial_account_id <> destination_financial_account_id)
		);
		CREATE TRIGGER transfers_immutable_delete BEFORE DELETE ON financial_transfers BEGIN SELECT RAISE(ABORT,'posted transfers cannot be deleted'); END;
		CREATE INDEX expenses_date ON expenses(expense_date DESC,expense_number DESC);
		CREATE INDEX transfers_date ON financial_transfers(transfer_date DESC,transfer_number DESC);`,
	},
	{
		version: 11,
		sql: `CREATE TABLE check_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL);
		INSERT INTO check_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE checks (
			id TEXT PRIMARY KEY, check_number TEXT NOT NULL UNIQUE, direction TEXT NOT NULL CHECK(direction IN ('incoming','outgoing')),
			bank TEXT NOT NULL, branch TEXT NOT NULL DEFAULT '', account_descriptor TEXT NOT NULL DEFAULT '', amount_rial INTEGER NOT NULL CHECK(amount_rial > 0),
			issue_date TEXT NOT NULL, due_date TEXT NOT NULL, payer_payee TEXT NOT NULL, customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL,
			supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL, source_type TEXT NOT NULL DEFAULT '', source_id TEXT NOT NULL DEFAULT '',
			financial_account_id TEXT REFERENCES financial_accounts(id) ON DELETE RESTRICT, notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL,
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
			CHECK((direction='incoming' AND status IN ('Draft','Received','Deposited','Cleared','Returned','Cancelled')) OR (direction='outgoing' AND status IN ('Draft','Issued','Delivered','Cleared','Returned','Rejected','Cancelled')))
		);
		CREATE INDEX checks_due ON checks(direction,status,due_date);
		CREATE TABLE check_events (
			id TEXT PRIMARY KEY, check_id TEXT NOT NULL REFERENCES checks(id) ON DELETE RESTRICT, from_status TEXT NOT NULL, to_status TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '', journal_entry_id TEXT NOT NULL DEFAULT '', idempotency_key TEXT NOT NULL UNIQUE, occurred_at TEXT NOT NULL
		);
		CREATE INDEX check_events_history ON check_events(check_id,occurred_at,id);
		CREATE TRIGGER checks_immutable_delete BEFORE DELETE ON checks WHEN OLD.status <> 'Draft' BEGIN SELECT RAISE(ABORT,'financial check history cannot be deleted'); END;
		CREATE TABLE loan_number_sequences (id INTEGER PRIMARY KEY CHECK(id=1), next_number INTEGER NOT NULL);
		INSERT INTO loan_number_sequences(id,next_number) VALUES(1,1001);
		CREATE TABLE loans (
			id TEXT PRIMARY KEY, loan_number TEXT NOT NULL UNIQUE, direction TEXT NOT NULL CHECK(direction IN ('payable','receivable')), counterparty_name TEXT NOT NULL,
			customer_id TEXT REFERENCES customers(id) ON DELETE SET NULL, supplier_id TEXT REFERENCES suppliers(id) ON DELETE SET NULL, principal_rial INTEGER NOT NULL CHECK(principal_rial > 0),
			interest_fee_rial INTEGER NOT NULL CHECK(interest_fee_rial >= 0), start_date TEXT NOT NULL, end_date TEXT, status TEXT NOT NULL CHECK(status IN ('Draft','Active','Closed','Cancelled')),
			notes TEXT NOT NULL DEFAULT '', financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT, journal_entry_id TEXT NOT NULL UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT,
			idempotency_key TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		);
		CREATE INDEX loans_status ON loans(status,direction,start_date);
		CREATE TABLE loan_installments (
			id TEXT PRIMARY KEY, loan_id TEXT NOT NULL REFERENCES loans(id) ON DELETE RESTRICT, position INTEGER NOT NULL, due_date TEXT NOT NULL,
			principal_rial INTEGER NOT NULL CHECK(principal_rial >= 0), interest_fee_rial INTEGER NOT NULL CHECK(interest_fee_rial >= 0), total_due_rial INTEGER NOT NULL CHECK(total_due_rial=principal_rial+interest_fee_rial),
			status TEXT NOT NULL DEFAULT 'Open' CHECK(status IN ('Open','Partially Paid','Paid')), UNIQUE(loan_id,position)
		);
		CREATE TABLE loan_payments (
			id TEXT PRIMARY KEY, payment_number TEXT NOT NULL UNIQUE, loan_id TEXT NOT NULL REFERENCES loans(id) ON DELETE RESTRICT, financial_account_id TEXT NOT NULL REFERENCES financial_accounts(id) ON DELETE RESTRICT,
			amount_rial INTEGER NOT NULL CHECK(amount_rial > 0), principal_rial INTEGER NOT NULL CHECK(principal_rial >= 0), interest_rial INTEGER NOT NULL CHECK(interest_rial >= 0),
			paid_at TEXT NOT NULL, notes TEXT NOT NULL DEFAULT '', status TEXT NOT NULL CHECK(status IN ('posted','reversed')), journal_entry_id TEXT NOT NULL UNIQUE REFERENCES journal_entries(id) ON DELETE RESTRICT,
			idempotency_key TEXT NOT NULL UNIQUE, created_at TEXT NOT NULL, CHECK(amount_rial=principal_rial+interest_rial)
		);
		CREATE TABLE loan_payment_allocations (
			id TEXT PRIMARY KEY, payment_id TEXT NOT NULL REFERENCES loan_payments(id) ON DELETE RESTRICT, installment_id TEXT NOT NULL REFERENCES loan_installments(id) ON DELETE RESTRICT,
			position INTEGER NOT NULL, principal_rial INTEGER NOT NULL CHECK(principal_rial >= 0), interest_rial INTEGER NOT NULL CHECK(interest_rial >= 0), UNIQUE(payment_id,position)
		);
		CREATE INDEX loan_installments_due ON loan_installments(due_date,status);`,
	},
}

func (s *Store) seedAccounting(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	accounts := []struct{ id, code, name, typ string }{
		{"ACC-CASH", "1000", "Cash", "asset"}, {"ACC-BANK", "1010", "Bank", "asset"},
		{"ACC-AR", "1100", "Accounts Receivable", "asset"}, {"ACC-INVENTORY", "1200", "Inventory", "asset"},
		{"ACC-OTHER-RECEIVABLE", "1300", "Other Receivables", "asset"}, {"ACC-AP", "2000", "Accounts Payable", "liability"},
		{"ACC-CUSTOMER-CREDIT", "2100", "Customer Credits", "liability"}, {"ACC-EQUITY", "3000", "Owner Equity", "equity"},
		{"ACC-REVENUE", "4000", "Sales Revenue (future invoice slice)", "revenue"}, {"ACC-COGS", "5000", "Cost of Goods Sold (future invoice slice)", "expense"},
		{"ACC-EXPENSE", "6000", "Operating Expenses", "expense"},
		{"ACC-EXP-RENT", "6100", "Rent", "expense"}, {"ACC-EXP-UTILITIES", "6110", "Electricity and Utilities", "expense"},
		{"ACC-EXP-INTERNET", "6120", "Internet", "expense"}, {"ACC-EXP-REPAIRS", "6130", "Repairs and Maintenance", "expense"},
		{"ACC-EXP-TRANSPORT", "6140", "Transport", "expense"}, {"ACC-EXP-SOFTWARE", "6150", "Software", "expense"},
		{"ACC-EXP-SALARIES", "6160", "Salaries and Wages (placeholder)", "expense"}, {"ACC-EXP-TAX", "6170", "Tax and Fees", "expense"},
		{"ACC-EXP-OTHER", "6190", "Other Expense", "expense"},
		{"ACC-CHECKS-RECEIVABLE", "1110", "Checks Receivable", "asset"}, {"ACC-CHECKS-IN-TRANSIT", "1120", "Checks in Transit", "asset"},
		{"ACC-LOANS-RECEIVABLE", "1400", "Loans Receivable", "asset"}, {"ACC-CHECKS-PAYABLE", "2010", "Checks Payable", "liability"},
		{"ACC-LOANS-PAYABLE", "2200", "Loans Payable", "liability"}, {"ACC-INTEREST-INCOME", "4100", "Interest Income", "revenue"},
		{"ACC-FINANCE-EXPENSE", "6200", "Finance Expense", "expense"},
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	fail := func(e error) error { _ = tx.Rollback(); return e }
	for _, a := range accounts {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO accounts(id,code,name,type,active,system,created_at,updated_at) VALUES(?,?,?,?,1,1,?,?)`, a.id, a.code, a.name, a.typ, now, now); err != nil {
			return fail(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO financial_accounts(id,name,type,ledger_account_id,details,active,created_at,updated_at) VALUES('FIN-CASH','Cash','cash','ACC-CASH','Default cash account',1,?,?)`, now, now); err != nil {
		return fail(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO financial_accounts(id,name,type,ledger_account_id,details,active,created_at,updated_at) VALUES('FIN-BANK','Bank','bank','ACC-BANK','Default bank account',1,?,?)`, now, now); err != nil {
		return fail(err)
	}
	return tx.Commit()
}

func (s *Store) ensureLegacyOpeningMovements(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, physical_stock_units, average_unit_cost_rial, updated_at FROM materials WHERE physical_stock_units > 0 AND NOT EXISTS (SELECT 1 FROM inventory_movements m WHERE m.material_id = materials.id AND m.movement_type = 'opening_balance')`)
	if err != nil {
		return fmt.Errorf("find legacy inventory: %w", err)
	}
	type opening struct {
		id        string
		qty, cost int64
		at        string
	}
	var openings []opening
	for rows.Next() {
		var o opening
		if err := rows.Scan(&o.id, &o.qty, &o.cost, &o.at); err != nil {
			return err
		}
		openings = append(openings, o)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close legacy inventory: %w", err)
	}
	if len(openings) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func(e error) error { _ = tx.Rollback(); return e }
	for _, o := range openings {
		total, e := domain.MulQuantityRial(domain.Quantity(o.qty), o.cost)
		if e != nil {
			return rollback(e)
		}
		if _, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-OPEN-"+o.id, o.id, o.at, "opening_balance", o.qty, o.cost, total, "material", o.id, "Migrated v6 opening balance", o.at); e != nil {
			return rollback(fmt.Errorf("backfill opening balance: %w", e))
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit opening balances: %w", err)
	}
	return nil
}

func (s *Store) List(ctx context.Context, includeArchived bool) ([]domain.Material, error) {
	query := `SELECT id, name, sku, category, purchase_unit, consumption_unit,
		conversion_factor_units, physical_stock_units, reorder_level_units,
		average_unit_cost_rial, preferred_supplier, notes, active, created_at, updated_at
		FROM materials`
	args := []any{}
	if !includeArchived {
		query += ` WHERE active = 1`
	}
	query += ` ORDER BY lower(name), id`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list materials: %w", err)
	}
	materials := []domain.Material{}
	for rows.Next() {
		material, err := scanMaterial(rows)
		if err != nil {
			return nil, fmt.Errorf("scan material: %w", err)
		}
		materials = append(materials, material)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read materials: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close materials: %w", err)
	}
	for i := range materials {
		materials[i], err = s.withInventorySummary(ctx, materials[i])
		if err != nil {
			return nil, err
		}
	}
	return materials, nil
}

func (s *Store) Get(ctx context.Context, id string) (domain.Material, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, sku, category, purchase_unit, consumption_unit,
		conversion_factor_units, physical_stock_units, reorder_level_units,
		average_unit_cost_rial, preferred_supplier, notes, active, created_at, updated_at
		FROM materials WHERE id = ?`, id)
	material, err := scanMaterial(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Material{}, domain.ErrMaterialNotFound
	}
	if err != nil {
		return domain.Material{}, fmt.Errorf("get material: %w", err)
	}
	material, err = s.withInventorySummary(ctx, material)
	if err != nil {
		return domain.Material{}, err
	}
	return material, nil
}

func (s *Store) Create(ctx context.Context, material domain.Material) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create material: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO materials
		(id, name, sku, category, purchase_unit, consumption_unit, conversion_factor_units,
		physical_stock_units, reorder_level_units, average_unit_cost_rial, preferred_supplier,
		notes, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		material.ID, material.Name, material.SKU, material.Category, material.PurchaseUnit, material.ConsumptionUnit,
		material.ConversionFactor, material.PhysicalStock, material.ReorderLevel, material.AverageUnitCostRial,
		material.PreferredSupplier, material.Notes, boolToInt(material.Active), material.CreatedAt.UTC().Format(time.RFC3339Nano), material.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("create material: %w", err)
	}
	if material.PhysicalStock > 0 {
		total, e := domain.MulQuantityRial(material.PhysicalStock, material.AverageUnitCostRial)
		if e != nil {
			_ = tx.Rollback()
			return e
		}
		_, e = tx.ExecContext(ctx, `INSERT INTO inventory_movements(id,material_id,occurred_at,movement_type,quantity_delta_units,unit_cost_rial,total_cost_rial,reference_type,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, "MOV-OPEN-"+material.ID, material.ID, material.CreatedAt.UTC().Format(time.RFC3339Nano), "opening_balance", material.PhysicalStock, material.AverageUnitCostRial, total, "material", material.ID, "Opening balance", material.CreatedAt.UTC().Format(time.RFC3339Nano))
		if e != nil {
			_ = tx.Rollback()
			return fmt.Errorf("create opening balance: %w", e)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit material: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, material domain.Material) error {
	result, err := s.db.ExecContext(ctx, `UPDATE materials SET name = ?, sku = ?, category = ?,
		purchase_unit = ?, consumption_unit = ?, conversion_factor_units = ?,
		reorder_level_units = ?, preferred_supplier = ?, notes = ?,
		active = ?, updated_at = ? WHERE id = ?`,
		material.Name, material.SKU, material.Category, material.PurchaseUnit, material.ConsumptionUnit,
		material.ConversionFactor, material.ReorderLevel,
		material.PreferredSupplier, material.Notes, boolToInt(material.Active), material.UpdatedAt.UTC().Format(time.RFC3339Nano), material.ID)
	if err != nil {
		return fmt.Errorf("update material: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check material update: %w", err)
	}
	if count == 0 {
		return domain.ErrMaterialNotFound
	}
	return nil
}

func (s *Store) withInventorySummary(ctx context.Context, material domain.Material) (domain.Material, error) {
	summary, err := s.inventorySummary(ctx, material.ID)
	if err != nil {
		return domain.Material{}, err
	}
	material.PhysicalStock = summary.PhysicalStock
	material.ReservedStock = summary.ReservedStock
	material.AvailableStock = summary.AvailableStock
	material.AverageUnitCostRial = summary.AverageUnitCostRial
	return material, nil
}

func (s *Store) inventorySummary(ctx context.Context, materialID string) (domain.InventorySummary, error) {
	return s.inventoryState(ctx, materialID)
}

func (s *Store) InventoryValue(ctx context.Context, materialID string) (int64, error) {
	var value int64
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_cost_rial),0) FROM inventory_movements WHERE material_id = ?`, materialID).Scan(&value)
	return value, err
}

func (s *Store) ListMachines(ctx context.Context, includeArchived bool) ([]domain.Machine, error) {
	query := `SELECT id, name, code, category, rate_basis, rate_rial, setup_cost_rial, notes, active, created_at, updated_at FROM machines`
	if !includeArchived {
		query += ` WHERE active = 1`
	}
	query += ` ORDER BY lower(name), id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}
	defer rows.Close()
	machines := []domain.Machine{}
	for rows.Next() {
		machine, scanErr := scanMachine(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan machine: %w", scanErr)
		}
		machines = append(machines, machine)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read machines: %w", err)
	}
	return machines, nil
}

func (s *Store) GetMachine(ctx context.Context, id string) (domain.Machine, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, code, category, rate_basis, rate_rial, setup_cost_rial, notes, active, created_at, updated_at FROM machines WHERE id = ?`, id)
	machine, err := scanMachine(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Machine{}, domain.ErrMachineNotFound
	}
	if err != nil {
		return domain.Machine{}, fmt.Errorf("get machine: %w", err)
	}
	return machine, nil
}

func (s *Store) SaveMachine(ctx context.Context, machine domain.Machine) error {
	result, err := s.db.ExecContext(ctx, `UPDATE machines SET name = ?, code = ?, category = ?, rate_basis = ?, rate_rial = ?, setup_cost_rial = ?, notes = ?, active = ?, updated_at = ? WHERE id = ?`, machine.Name, machine.Code, machine.Category, machine.RateBasis, machine.RateRial, machine.SetupCostRial, machine.Notes, boolToInt(machine.Active), machine.UpdatedAt.UTC().Format(time.RFC3339Nano), machine.ID)
	if err != nil {
		return fmt.Errorf("update machine: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check machine update: %w", err)
	}
	if count == 0 {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO machines (id, name, code, category, rate_basis, rate_rial, setup_cost_rial, notes, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, machine.ID, machine.Name, machine.Code, machine.Category, machine.RateBasis, machine.RateRial, machine.SetupCostRial, machine.Notes, boolToInt(machine.Active), machine.CreatedAt.UTC().Format(time.RFC3339Nano), machine.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("insert machine: %w", err)
		}
	}
	return nil
}

func (s *Store) ListServices(ctx context.Context, includeArchived bool) ([]domain.Service, error) {
	query := `SELECT id, name, code, category, description, active, created_at, updated_at FROM services`
	if !includeArchived {
		query += ` WHERE active = 1`
	}
	query += ` ORDER BY lower(name), id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	services := []domain.Service{}
	for rows.Next() {
		service, err := scanService(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("read services: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close services: %w", err)
	}
	for index := range services {
		services[index].Parameters, err = s.loadParameters(ctx, services[index].ID)
		if err != nil {
			return nil, err
		}
		services[index].Components, err = s.loadComponents(ctx, services[index].ID)
		if err != nil {
			return nil, err
		}
		services[index].PricingRule, err = s.loadPricingRule(ctx, services[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return services, nil
}

func (s *Store) GetService(ctx context.Context, id string) (domain.Service, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, code, category, description, active, created_at, updated_at FROM services WHERE id = ?`, id)
	service, err := scanService(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Service{}, domain.ErrServiceNotFound
	}
	if err != nil {
		return domain.Service{}, fmt.Errorf("get service: %w", err)
	}
	service.Parameters, err = s.loadParameters(ctx, service.ID)
	if err != nil {
		return domain.Service{}, err
	}
	service.Components, err = s.loadComponents(ctx, service.ID)
	if err != nil {
		return domain.Service{}, err
	}
	service.PricingRule, err = s.loadPricingRule(ctx, service.ID)
	if err != nil {
		return domain.Service{}, err
	}
	return service, nil
}

func (s *Store) loadParameters(ctx context.Context, serviceID string) ([]domain.ServiceParameter, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, service_id, parameter_key, label, parameter_type, required,
		display_order, default_value, min_value_units, max_value_units, unit_label, active, created_at, updated_at
		FROM service_parameters WHERE service_id = ? ORDER BY display_order, id`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list service parameters: %w", err)
	}
	parameters := []domain.ServiceParameter{}
	for rows.Next() {
		parameter, err := scanParameter(rows)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan service parameter: %w", err)
		}
		parameters = append(parameters, parameter)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("read service parameters: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close service parameters: %w", err)
	}
	for index := range parameters {
		parameters[index].Options, err = s.loadParameterOptions(ctx, parameters[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return parameters, nil
}

func (s *Store) loadParameterOptions(ctx context.Context, parameterID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT value FROM service_parameter_options WHERE parameter_id = ? ORDER BY option_order`, parameterID)
	if err != nil {
		return nil, fmt.Errorf("list parameter options: %w", err)
	}
	defer rows.Close()
	options := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, fmt.Errorf("scan parameter option: %w", err)
		}
		options = append(options, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read parameter options: %w", err)
	}
	return options, nil
}

func (s *Store) SaveServiceDefinition(ctx context.Context, service domain.Service) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin service definition write: %w", err)
	}
	rollback := func(writeErr error) error {
		_ = tx.Rollback()
		return writeErr
	}
	result, err := tx.ExecContext(ctx, `UPDATE services SET name = ?, code = ?, category = ?, description = ?, active = ?, updated_at = ? WHERE id = ?`,
		service.Name, service.Code, service.Category, service.Description, boolToInt(service.Active), service.UpdatedAt.UTC().Format(time.RFC3339Nano), service.ID)
	if err != nil {
		return rollback(fmt.Errorf("update service: %w", err))
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return rollback(fmt.Errorf("check service update: %w", err))
	}
	if updated == 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO services (id, name, code, category, description, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			service.ID, service.Name, service.Code, service.Category, service.Description, boolToInt(service.Active), service.CreatedAt.UTC().Format(time.RFC3339Nano), service.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return rollback(fmt.Errorf("insert service: %w", err))
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM service_parameters WHERE service_id = ?`, service.ID); err != nil {
		return rollback(fmt.Errorf("replace service parameters: %w", err))
	}
	for _, parameter := range service.Parameters {
		var minimum any
		if parameter.MinValue != nil {
			minimum = int64(*parameter.MinValue)
		}
		var maximum any
		if parameter.MaxValue != nil {
			maximum = int64(*parameter.MaxValue)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO service_parameters
			(id, service_id, parameter_key, label, parameter_type, required, display_order, default_value,
			min_value_units, max_value_units, unit_label, active, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			parameter.ID, service.ID, parameter.Key, parameter.Label, string(parameter.Type), boolToInt(parameter.Required), parameter.Position,
			parameter.DefaultValue, minimum, maximum, parameter.Unit, boolToInt(parameter.Active), parameter.CreatedAt.UTC().Format(time.RFC3339Nano), parameter.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return rollback(fmt.Errorf("insert service parameter: %w", err))
		}
		for optionOrder, option := range parameter.Options {
			if _, err := tx.ExecContext(ctx, `INSERT INTO service_parameter_options (parameter_id, option_order, value) VALUES (?, ?, ?)`, parameter.ID, optionOrder, option); err != nil {
				return rollback(fmt.Errorf("insert parameter option: %w", err))
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM service_cost_components WHERE service_id = ?`, service.ID); err != nil {
		return rollback(fmt.Errorf("replace service cost components: %w", err))
	}
	for _, component := range service.Components {
		if _, err := tx.ExecContext(ctx, `INSERT INTO service_cost_components (id, service_id, component_name, component_type, reference_id, usage_mode, parameter_key, usage_quantity_units, multiplier_units, rate_rial, percentage_units, rate_basis, enabled, display_order, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, component.ID, service.ID, component.Name, string(component.Type), component.ReferenceID, string(component.UsageMode), component.ParameterKey, int64(component.UsageQuantity), int64(component.Multiplier), component.RateRial, int64(component.Percentage), component.RateBasis, boolToInt(component.Enabled), component.Position, component.Notes, component.CreatedAt.UTC().Format(time.RFC3339Nano), component.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return rollback(fmt.Errorf("insert service cost component: %w", err))
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM service_pricing_rules WHERE service_id = ?`, service.ID); err != nil {
		return rollback(fmt.Errorf("replace service pricing rule: %w", err))
	}
	if service.PricingRule != nil {
		rule := service.PricingRule
		if _, err := tx.ExecContext(ctx, `INSERT INTO service_pricing_rules (id, service_id, rule_type, fixed_price_rial, markup_percentage_units, fixed_margin_rial, per_unit_rate_rial, parameter_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, rule.ID, service.ID, string(rule.Type), rule.FixedPriceRial, int64(rule.MarkupPercentage), rule.FixedMarginRial, rule.PerUnitRateRial, rule.ParameterKey, rule.CreatedAt.UTC().Format(time.RFC3339Nano), rule.UpdatedAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return rollback(fmt.Errorf("insert service pricing rule: %w", err))
		}
		for _, tier := range rule.Tiers {
			if _, err := tx.ExecContext(ctx, `INSERT INTO service_pricing_tiers (rule_id, display_order, minimum_quantity_units, price_rial) VALUES (?, ?, ?, ?)`, rule.ID, tier.Position, int64(tier.MinimumQuantity), tier.PriceRial); err != nil {
				return rollback(fmt.Errorf("insert service pricing tier: %w", err))
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit service definition: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(...any) error
}

func scanMaterial(row scanner) (domain.Material, error) {
	var material domain.Material
	var conversion, physical, reorder int64
	var active int
	var created, updated string
	err := row.Scan(&material.ID, &material.Name, &material.SKU, &material.Category, &material.PurchaseUnit, &material.ConsumptionUnit,
		&conversion, &physical, &reorder, &material.AverageUnitCostRial, &material.PreferredSupplier, &material.Notes,
		&active, &created, &updated)
	if err != nil {
		return domain.Material{}, err
	}
	material.ConversionFactor = domain.Quantity(conversion)
	material.PhysicalStock = domain.Quantity(physical)
	material.ReorderLevel = domain.Quantity(reorder)
	material.Active = active == 1
	material.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return domain.Material{}, fmt.Errorf("parse created timestamp: %w", err)
	}
	material.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return domain.Material{}, fmt.Errorf("parse updated timestamp: %w", err)
	}
	return material, nil
}

func scanService(row scanner) (domain.Service, error) {
	var service domain.Service
	var active int
	var created, updated string
	if err := row.Scan(&service.ID, &service.Name, &service.Code, &service.Category, &service.Description, &active, &created, &updated); err != nil {
		return domain.Service{}, err
	}
	service.Active = active == 1
	var err error
	service.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return domain.Service{}, fmt.Errorf("parse service created timestamp: %w", err)
	}
	service.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return domain.Service{}, fmt.Errorf("parse service updated timestamp: %w", err)
	}
	return service, nil
}

func scanParameter(row scanner) (domain.ServiceParameter, error) {
	var parameter domain.ServiceParameter
	var parameterType string
	var required, active int
	var minimum, maximum sql.NullInt64
	var created, updated string
	if err := row.Scan(&parameter.ID, &parameter.ServiceID, &parameter.Key, &parameter.Label, &parameterType, &required,
		&parameter.Position, &parameter.DefaultValue, &minimum, &maximum, &parameter.Unit, &active, &created, &updated); err != nil {
		return domain.ServiceParameter{}, err
	}
	parameter.Type = domain.ParameterType(parameterType)
	parameter.Required = required == 1
	parameter.Active = active == 1
	if minimum.Valid {
		value := domain.Quantity(minimum.Int64)
		parameter.MinValue = &value
	}
	if maximum.Valid {
		value := domain.Quantity(maximum.Int64)
		parameter.MaxValue = &value
	}
	var err error
	parameter.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return domain.ServiceParameter{}, fmt.Errorf("parse parameter created timestamp: %w", err)
	}
	parameter.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return domain.ServiceParameter{}, fmt.Errorf("parse parameter updated timestamp: %w", err)
	}
	return parameter, nil
}

func (s *Store) loadComponents(ctx context.Context, serviceID string) ([]domain.ServiceCostComponent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, service_id, component_name, component_type, reference_id, usage_mode, parameter_key, usage_quantity_units, multiplier_units, rate_rial, percentage_units, rate_basis, enabled, display_order, notes, created_at, updated_at FROM service_cost_components WHERE service_id = ? ORDER BY display_order, id`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("list service cost components: %w", err)
	}
	defer rows.Close()
	components := []domain.ServiceCostComponent{}
	for rows.Next() {
		component, scanErr := scanComponent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan service cost component: %w", scanErr)
		}
		components = append(components, component)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read service cost components: %w", err)
	}
	return components, nil
}

func scanMachine(row scanner) (domain.Machine, error) {
	var machine domain.Machine
	var active int
	var created, updated string
	if err := row.Scan(&machine.ID, &machine.Name, &machine.Code, &machine.Category, &machine.RateBasis, &machine.RateRial, &machine.SetupCostRial, &machine.Notes, &active, &created, &updated); err != nil {
		return domain.Machine{}, err
	}
	machine.Active = active == 1
	var err error
	machine.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return domain.Machine{}, fmt.Errorf("parse machine created timestamp: %w", err)
	}
	machine.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return domain.Machine{}, fmt.Errorf("parse machine updated timestamp: %w", err)
	}
	return machine, nil
}

func scanComponent(row scanner) (domain.ServiceCostComponent, error) {
	var component domain.ServiceCostComponent
	var componentType, usageMode string
	var usageQuantity, multiplier, percentage int64
	var enabled int
	var created, updated string
	if err := row.Scan(&component.ID, &component.ServiceID, &component.Name, &componentType, &component.ReferenceID, &usageMode, &component.ParameterKey, &usageQuantity, &multiplier, &component.RateRial, &percentage, &component.RateBasis, &enabled, &component.Position, &component.Notes, &created, &updated); err != nil {
		return domain.ServiceCostComponent{}, err
	}
	component.Type = domain.CostComponentType(componentType)
	component.UsageMode = domain.UsageMode(usageMode)
	component.UsageQuantity = domain.Quantity(usageQuantity)
	component.Multiplier = domain.Quantity(multiplier)
	component.Percentage = domain.Quantity(percentage)
	component.Enabled = enabled == 1
	var err error
	component.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return domain.ServiceCostComponent{}, fmt.Errorf("parse component created timestamp: %w", err)
	}
	component.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return domain.ServiceCostComponent{}, fmt.Errorf("parse component updated timestamp: %w", err)
	}
	return component, nil
}

func (s *Store) loadPricingRule(ctx context.Context, serviceID string) (*domain.ServicePricingRule, error) {
	var rule domain.ServicePricingRule
	var ruleType, created, updated string
	var markup int64
	err := s.db.QueryRowContext(ctx, `SELECT id, service_id, rule_type, fixed_price_rial, markup_percentage_units, fixed_margin_rial, per_unit_rate_rial, parameter_key, created_at, updated_at FROM service_pricing_rules WHERE service_id = ?`, serviceID).Scan(&rule.ID, &rule.ServiceID, &ruleType, &rule.FixedPriceRial, &markup, &rule.FixedMarginRial, &rule.PerUnitRateRial, &rule.ParameterKey, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get service pricing rule: %w", err)
	}
	rule.Type = domain.PricingRuleType(ruleType)
	rule.MarkupPercentage = domain.Quantity(markup)
	rule.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return nil, fmt.Errorf("parse pricing rule created timestamp: %w", err)
	}
	rule.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return nil, fmt.Errorf("parse pricing rule updated timestamp: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT display_order, minimum_quantity_units, price_rial FROM service_pricing_tiers WHERE rule_id = ? ORDER BY display_order`, rule.ID)
	if err != nil {
		return nil, fmt.Errorf("list pricing tiers: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var position, minimum int64
		var price int64
		if err := rows.Scan(&position, &minimum, &price); err != nil {
			return nil, fmt.Errorf("scan pricing tier: %w", err)
		}
		rule.Tiers = append(rule.Tiers, domain.ServicePricingTier{Position: int(position), MinimumQuantity: domain.Quantity(minimum), PriceRial: price})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read pricing tiers: %w", err)
	}
	return &rule, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (s *Store) ListCustomers(ctx context.Context, includeArchived bool) ([]domain.Customer, error) {
	query := `SELECT id,name,phone,email,address,notes,active,created_at,updated_at FROM customers`
	if !includeArchived {
		query += ` WHERE active = 1`
	}
	query += ` ORDER BY lower(name), id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	defer rows.Close()
	var result []domain.Customer
	for rows.Next() {
		item, scanErr := scanCustomer(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read customers: %w", err)
	}
	return result, nil
}

func (s *Store) GetCustomer(ctx context.Context, id string) (domain.Customer, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,name,phone,email,address,notes,active,created_at,updated_at FROM customers WHERE id = ?`, id)
	c, err := scanCustomer(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Customer{}, domain.ErrCustomerNotFound
	}
	if err != nil {
		return domain.Customer{}, fmt.Errorf("get customer: %w", err)
	}
	return c, nil
}

func (s *Store) SaveCustomer(ctx context.Context, customer domain.Customer) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO customers(id,name,phone,email,address,notes,active,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name,phone=excluded.phone,email=excluded.email,address=excluded.address,notes=excluded.notes,active=excluded.active,updated_at=excluded.updated_at`,
		customer.ID, customer.Name, customer.Phone, customer.Email, customer.Address, customer.Notes, boolToInt(customer.Active), customer.CreatedAt.UTC().Format(time.RFC3339Nano), customer.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("save customer: %w", err)
	}
	return nil
}

func (s *Store) ListOrders(ctx context.Context) ([]domain.Order, error) {
	rows, err := s.db.QueryContext(ctx, orderSelect+` ORDER BY created_at DESC, order_number DESC`)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()
	var result []domain.Order
	for rows.Next() {
		order, scanErr := scanOrder(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if err := s.loadOrderItems(ctx, &order); err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read orders: %w", err)
	}
	return result, nil
}

const orderSelect = `SELECT id,order_number,customer_id,customer_name_snapshot,customer_phone_snapshot,created_at,promised_at,priority,commercial_status,fulfillment_status,payment_status,notes,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at,quote_id FROM orders`

func (s *Store) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	row := s.db.QueryRowContext(ctx, orderSelect+` WHERE id = ?`, id)
	order, err := scanOrder(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}
	if err := s.loadOrderItems(ctx, &order); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}

func (s *Store) CreateOrder(ctx context.Context, order domain.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin order create: %w", err)
	}
	rollback := func(e error) error { _ = tx.Rollback(); return e }
	if _, err := tx.ExecContext(ctx, `UPDATE order_number_sequences SET next_number = next_number + 1 WHERE id = 1`); err != nil {
		return rollback(fmt.Errorf("advance order number: %w", err))
	}
	var number int64
	if err := tx.QueryRowContext(ctx, `SELECT next_number - 1 FROM order_number_sequences WHERE id = 1`).Scan(&number); err != nil {
		return rollback(fmt.Errorf("read order number: %w", err))
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
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit order create: %w", err)
	}
	return nil
}

func (s *Store) SaveOrder(ctx context.Context, order domain.Order) error {
	if err := order.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin order save: %w", err)
	}
	rollback := func(e error) error { _ = tx.Rollback(); return e }
	var paid, total int64
	if err := tx.QueryRowContext(ctx, `SELECT total_rial FROM orders WHERE id=?`, order.ID).Scan(&total); err == nil {
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(a.amount_rial),0) FROM payment_allocations a JOIN payments p ON p.id=a.payment_id WHERE a.target_type='order' AND a.target_id=? AND a.reversed=0 AND p.status='posted'`, order.ID).Scan(&paid); err != nil {
			return rollback(err)
		}
		if paid <= 0 {
			order.PaymentStatus = domain.PaymentUnpaid
		} else if paid < total {
			order.PaymentStatus = domain.PaymentPartiallyPaid
		} else {
			order.PaymentStatus = domain.PaymentPaid
		}
	}
	result, err := tx.ExecContext(ctx, `UPDATE orders SET customer_id=?,customer_name_snapshot=?,customer_phone_snapshot=?,created_at=?,promised_at=?,priority=?,commercial_status=?,fulfillment_status=?,payment_status=?,notes=?,subtotal_rial=?,discount_rial=?,total_rial=?,estimated_cost_rial=?,updated_at=?,quote_id=? WHERE id=?`, nullableString(order.CustomerID), order.CustomerNameSnapshot, order.CustomerPhoneSnapshot, order.CreatedAt.UTC().Format(time.RFC3339Nano), nullableTime(order.PromisedAt), string(order.Priority), string(order.CommercialStatus), string(order.FulfillmentStatus), string(order.PaymentStatus), order.Notes, order.SubtotalRial, order.DiscountRial, order.TotalRial, order.EstimatedCostRial, order.UpdatedAt.UTC().Format(time.RFC3339Nano), nullableString(order.QuoteID), order.ID)
	if err != nil {
		return rollback(fmt.Errorf("update order: %w", err))
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return rollback(domain.ErrOrderNotFound)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE inventory_reservations SET status=CASE WHEN ?='Cancelled' THEN 'cancelled' ELSE 'released' END, updated_at=? WHERE order_id=? AND status='active' AND (?='Cancelled' OR ?='Draft')`, string(order.CommercialStatus), order.UpdatedAt.UTC().Format(time.RFC3339Nano), order.ID, string(order.CommercialStatus), string(order.CommercialStatus)); err != nil {
		return rollback(fmt.Errorf("release order reservations: %w", err))
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id=?`, order.ID); err != nil {
		return rollback(fmt.Errorf("replace order items: %w", err))
	}
	if err := insertOrderItems(ctx, tx, order); err != nil {
		return rollback(err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit order save: %w", err)
	}
	return nil
}

func insertOrder(ctx context.Context, tx *sql.Tx, order domain.Order) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO orders(id,order_number,customer_id,customer_name_snapshot,customer_phone_snapshot,created_at,promised_at,priority,commercial_status,fulfillment_status,payment_status,notes,subtotal_rial,discount_rial,total_rial,estimated_cost_rial,updated_at,quote_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, order.ID, order.OrderNumber, nullableString(order.CustomerID), order.CustomerNameSnapshot, order.CustomerPhoneSnapshot, order.CreatedAt.UTC().Format(time.RFC3339Nano), nullableTime(order.PromisedAt), string(order.Priority), string(order.CommercialStatus), string(order.FulfillmentStatus), string(order.PaymentStatus), order.Notes, order.SubtotalRial, order.DiscountRial, order.TotalRial, order.EstimatedCostRial, order.UpdatedAt.UTC().Format(time.RFC3339Nano), nullableString(order.QuoteID))
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	return nil
}
func insertOrderItems(ctx context.Context, tx *sql.Tx, order domain.Order) error {
	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, `INSERT INTO order_items(id,order_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, item.ID, order.ID, item.Position, item.ServiceID, item.ServiceNameSnapshot, item.ServiceCodeSnapshot, int64(item.Quantity), item.QuantityUnit, item.ResolvedParametersJSON, item.CostBreakdownJSON, item.PricingSnapshotJSON, item.EstimatedCostRial, item.SuggestedPriceRial, item.SellingPriceRial, item.Notes); err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}
	return nil
}

func (s *Store) loadOrderItems(ctx context.Context, order *domain.Order) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id,order_id,display_order,service_id,service_name_snapshot,service_code_snapshot,quantity_units,quantity_unit,resolved_parameters_json,cost_breakdown_json,pricing_snapshot_json,estimated_cost_rial,suggested_price_rial,selling_price_rial,notes FROM order_items WHERE order_id=? ORDER BY display_order,id`, order.ID)
	if err != nil {
		return fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		item, scanErr := scanOrderItem(rows)
		if scanErr != nil {
			return scanErr
		}
		order.Items = append(order.Items, item)
	}
	return rows.Err()
}
func scanCustomer(row scanner) (domain.Customer, error) {
	var c domain.Customer
	var active int
	var created, updated string
	if err := row.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Address, &c.Notes, &active, &created, &updated); err != nil {
		return c, err
	}
	c.Active = active == 1
	var err error
	c.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return c, err
	}
	c.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	return c, err
}
func scanOrder(row scanner) (domain.Order, error) {
	var o domain.Order
	var customerID sql.NullString
	var promised sql.NullString
	var priority, commercial, fulfillment, payment, created, updated string
	var quoteID sql.NullString
	if err := row.Scan(&o.ID, &o.OrderNumber, &customerID, &o.CustomerNameSnapshot, &o.CustomerPhoneSnapshot, &created, &promised, &priority, &commercial, &fulfillment, &payment, &o.Notes, &o.SubtotalRial, &o.DiscountRial, &o.TotalRial, &o.EstimatedCostRial, &updated, &quoteID); err != nil {
		return o, err
	}
	o.CustomerID = customerID.String
	o.QuoteID = quoteID.String
	o.Priority = domain.Priority(priority)
	o.CommercialStatus = domain.CommercialStatus(commercial)
	o.FulfillmentStatus = domain.FulfillmentStatus(fulfillment)
	o.PaymentStatus = domain.PaymentStatus(payment)
	var err error
	o.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return o, err
	}
	o.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return o, err
	}
	if promised.Valid {
		x, parseErr := time.Parse(time.RFC3339Nano, promised.String)
		if parseErr != nil {
			return o, parseErr
		}
		o.PromisedAt = &x
	}
	return o, nil
}
func scanOrderItem(row scanner) (domain.OrderItem, error) {
	var i domain.OrderItem
	var quantity int64
	if err := row.Scan(&i.ID, &i.OrderID, &i.Position, &i.ServiceID, &i.ServiceNameSnapshot, &i.ServiceCodeSnapshot, &quantity, &i.QuantityUnit, &i.ResolvedParametersJSON, &i.CostBreakdownJSON, &i.PricingSnapshotJSON, &i.EstimatedCostRial, &i.SuggestedPriceRial, &i.SellingPriceRial, &i.Notes); err != nil {
		return i, err
	}
	i.Quantity = domain.Quantity(quantity)
	return i, nil
}
func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}
