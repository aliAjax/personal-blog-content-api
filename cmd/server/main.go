package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/blog-api/internal/router"
	"github.com/example/blog-api/internal/seed"
	"github.com/example/blog-api/pkg/config"
	"github.com/example/blog-api/pkg/database"
	"github.com/example/blog-api/pkg/logger"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		logger.Error("database open failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	migrationCtx, cancelMigration := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelMigration()
	if err := database.Migrate(migrationCtx, db, cfg.MigrationsDir); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	seedCtx, cancelSeed := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelSeed()
	if err := seed.Run(seedCtx, db, cfg); err != nil {
		logger.Error("seed failed", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(db, cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server listen failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
