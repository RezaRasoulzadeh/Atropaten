package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"Atropaten/internal/application"
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
	configDir, err := os.UserConfigDir()
	if err != nil {
		a.startupError = fmt.Errorf("find application data directory: %w", err)
		return
	}
	database, err := sqlite.Open(filepath.Join(configDir, "Atropaten", "atropaten.db"))
	if err != nil {
		a.startupError = err
		return
	}
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
}

func (a *App) shutdown(_ context.Context) {
	if a.database != nil {
		_ = a.database.Close()
	}
}
