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
		if err := application.RunMigrationCommand(ctx, os.Args[1], os.Args[2:]); err != nil {
			log.Fatalf("migration command failed: %v", err)
		}
		return
	}

	if err := application.RunMigrationCommand(ctx, "up", nil); err != nil {
		log.Fatalf("migration up failed: %v", err)
	}

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
