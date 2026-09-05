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

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
