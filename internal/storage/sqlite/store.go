package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	)`},
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
	defer rows.Close()
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
	return material, nil
}

func (s *Store) Create(ctx context.Context, material domain.Material) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO materials
		(id, name, sku, category, purchase_unit, consumption_unit, conversion_factor_units,
		physical_stock_units, reorder_level_units, average_unit_cost_rial, preferred_supplier,
		notes, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		material.ID, material.Name, material.SKU, material.Category, material.PurchaseUnit, material.ConsumptionUnit,
		material.ConversionFactor, material.PhysicalStock, material.ReorderLevel, material.AverageUnitCostRial,
		material.PreferredSupplier, material.Notes, boolToInt(material.Active), material.CreatedAt.UTC().Format(time.RFC3339Nano), material.UpdatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("create material: %w", err)
	}
	return nil
}

func (s *Store) Update(ctx context.Context, material domain.Material) error {
	result, err := s.db.ExecContext(ctx, `UPDATE materials SET name = ?, sku = ?, category = ?,
		purchase_unit = ?, consumption_unit = ?, conversion_factor_units = ?, physical_stock_units = ?,
		reorder_level_units = ?, average_unit_cost_rial = ?, preferred_supplier = ?, notes = ?,
		active = ?, updated_at = ? WHERE id = ?`,
		material.Name, material.SKU, material.Category, material.PurchaseUnit, material.ConsumptionUnit,
		material.ConversionFactor, material.PhysicalStock, material.ReorderLevel, material.AverageUnitCostRial,
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

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
