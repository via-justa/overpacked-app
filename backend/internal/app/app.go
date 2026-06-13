package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/via-justa/overpacked-app/backend/internal/auth"
	"github.com/via-justa/overpacked-app/backend/internal/backup"
	"github.com/via-justa/overpacked-app/backend/internal/config"
	"github.com/via-justa/overpacked-app/backend/internal/db"
	"github.com/via-justa/overpacked-app/backend/internal/http/handlers"
	"github.com/via-justa/overpacked-app/backend/internal/migrations"
	"github.com/via-justa/overpacked-app/backend/internal/seeds"
	"github.com/via-justa/overpacked-app/backend/internal/store"
)

type App struct {
	cfg       *config.Config
	db        *sql.DB
	server    *http.Server
	Store     *store.Store
	scheduler *backup.Scheduler
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	if err := database.PingContext(ctx); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	authService, err := auth.NewService(cfg.AppUsername, cfg.AppPassword, cfg.JWTSecret)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("init auth service: %w", err)
	}
	authHandler := handlers.NewAuthHandler(authService)
	st := store.New(database)

	backupService := backup.NewService(database, cfg.BackupBaseDir)
	scheduler := backup.NewScheduler(backupService, st.BackupConfig)
	backupHandler := handlers.NewBackupHandler(backupService, st, scheduler, cfg.AppPassword)

	router := chi.NewRouter()
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.Logger)
	router.Use(securityHeaders)
	router.Use(limitBody)
	router.Use(requireAuth(authService))
	router.Mount("/", setupRoutes(authHandler, st, cfg.AppPassword, backupHandler))

	app := &App{
		cfg:       cfg,
		db:        database,
		Store:     st,
		scheduler: scheduler,
		server: &http.Server{
			Addr:              cfg.ServerAddr,
			Handler:           router,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}

	return app, nil
}

func (a *App) Start() error {
	fmt.Printf("starting server on %s\n", a.cfg.ServerAddr)
	return a.server.ListenAndServe()
}

// StartScheduler launches the background backup scheduler. Call only on the server
// path, not for one-shot CLI commands.
func (a *App) StartScheduler(ctx context.Context) error {
	return a.scheduler.Start(ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	a.scheduler.Stop()
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("close db: %w", err)
	}
	return nil
}

func (a *App) RunMigrationCommand(ctx context.Context, command string, args []string) error {
	return migrations.Run(ctx, a.db, command, args)
}

func (a *App) RunSeeds(ctx context.Context) error {
	return seeds.Run(ctx, a.db)
}
