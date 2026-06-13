package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/via-justa/overpacked-app/backend/internal/app"
	"github.com/via-justa/overpacked-app/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("app init: %v", err)
	}

	if len(os.Args) > 1 {
		if err := runCLICommand(ctx, application, os.Args[1:]); err != nil {
			log.Fatalf("%v", err)
		}
		return
	}

	if err := bootstrap(ctx, application, cfg); err != nil {
		log.Fatalf("%v", err)
	}

	runServer(application)
}

// runCLICommand handles one-shot subcommands (seed, migrations) and returns.
func runCLICommand(ctx context.Context, application *app.App, args []string) error {
	command := args[0]
	if command == "seed" {
		if err := application.RunSeeds(ctx); err != nil {
			return fmt.Errorf("seed command failed: %w", err)
		}
		return nil
	}
	if err := application.RunMigrationCommand(ctx, command, args[1:]); err != nil {
		return fmt.Errorf("migration command failed: %w", err)
	}
	return nil
}

// bootstrap runs startup migrations, optional seeding, and the backup scheduler.
func bootstrap(ctx context.Context, application *app.App, cfg *config.Config) error {
	if err := application.RunMigrationCommand(ctx, "up", nil); err != nil {
		return fmt.Errorf("migration up failed: %w", err)
	}
	if cfg.EnableSeedData {
		if err := application.RunSeeds(ctx); err != nil {
			return fmt.Errorf("seed on startup failed: %w", err)
		}
	}
	if err := application.StartScheduler(ctx); err != nil {
		return fmt.Errorf("start backup scheduler failed: %w", err)
	}
	return nil
}

// runServer starts the HTTP server and blocks until a termination signal, then
// shuts down gracefully.
func runServer(application *app.App) {
	go func() {
		if err := application.Start(); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nshutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	fmt.Println("server stopped")
}
