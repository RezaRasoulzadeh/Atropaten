package main

import (
	"context"
	"fmt"

	"Atropaten/internal/application"
	"Atropaten/internal/platform"
	"Atropaten/internal/storage/sqlite"
)

// App struct
type App struct {
	ctx          context.Context
	materials    *application.MaterialsService
	services     *application.ServicesService
	pricing      *application.PricingService
	machines     *application.MachinesService
	customers    *application.CustomersService
	orders       *application.OrdersService
	quotes       *application.QuotesService
	metadata     *application.MetadataService
	suppliers    *application.SuppliersService
	purchases    *application.PurchasesService
	production   *application.ProductionService
	accounting   *application.AccountingService
	invoices     *application.InvoicesService
	checks       *application.ChecksService
	loans        *application.LoansService
	owners       *application.OwnersService
	reporting    *application.ReportingService
	paths        platform.DataPaths
	backup       *platform.BackupService
	database     *sqlite.Store
	startupError error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	paths, err := platform.ResolveDataPaths("Atropaten")
	if err != nil {
		a.startupError = fmt.Errorf("resolve application data directory: %w", err)
		return
	}
	if err = paths.Ensure(); err != nil {
		a.startupError = fmt.Errorf("create application data directory: %w", err)
		return
	}
	database, err := sqlite.Open(paths.Database)
	if err != nil {
		a.startupError = err
		return
	}
	a.database = database
	a.paths = paths
	a.configureServices(database)
	a.backup = platform.NewBackupService(paths, database, sqlite.ValidateDatabaseFile)
}

func (a *App) configureServices(database *sqlite.Store) {
	a.database = database
	a.materials = application.NewMaterialsService(database)
	a.machines = application.NewMachinesService(database)
	a.services = application.NewServicesService(database, database, database)
	a.pricing = application.NewPricingService(database, database, database)
	a.customers = application.NewCustomersService(database)
	a.orders = application.NewOrdersService(database, database, a.pricing)
	a.quotes = application.NewQuotesService(database, database, a.pricing)
	a.metadata = application.NewMetadataService(database, database)
	a.suppliers = application.NewSuppliersService(database)
	a.purchases = application.NewPurchasesService(database, database, database)
	a.production = application.NewProductionService(database)
	a.accounting = application.NewAccountingService(database)
	a.invoices = application.NewInvoicesService(database, database)
	a.checks = application.NewChecksService(database)
	a.loans = application.NewLoansService(database)
	a.owners = application.NewOwnersService(database)
	a.reporting = application.NewReportingService(database)
}

func (a *App) closeForRestore() error {
	if a.database == nil {
		return nil
	}
	err := a.database.Close()
	a.database = nil
	return err
}

func (a *App) reopenAfterRestore() error {
	database, err := sqlite.Open(a.paths.Database)
	if err != nil {
		return err
	}
	a.configureServices(database)
	if a.backup != nil {
		a.backup.SetRepository(database)
	}
	return nil
}

func (a *App) shutdown(_ context.Context) {
	if a.database != nil {
		_ = a.database.Close()
	}
}
