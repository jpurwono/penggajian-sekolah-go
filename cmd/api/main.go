// API Penggajian Sekolah — entry point aplikasi.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"penggajian/server/internal/platform/config"
	"penggajian/server/internal/platform/database"
	"penggajian/server/internal/platform/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[config] %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[db] %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("[db] migrasi gagal: %v", err)
	}
	if err := database.Seed(db); err != nil {
		log.Fatalf("[db] seed gagal: %v", err)
	}

	e := router.New(cfg, db)

	// Graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		addr := ":" + cfg.AppPort
		log.Printf("[app] %s berjalan di http://localhost%s (env: %s)", cfg.AppName, addr, cfg.AppEnv)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[app] server gagal jalan: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("[app] mematikan server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("[app] shutdown error: %v", err)
	}
	log.Println("[app] berhenti")
}
