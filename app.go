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
}

func (a *App) shutdown(_ context.Context) {
	if a.database != nil {
		_ = a.database.Close()
	}
}
