package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"book-reader/backend/internal/app"
	"book-reader/backend/internal/config"
	"book-reader/backend/internal/database"
	"book-reader/backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	store := storage.NewLocal(cfg)
	if err := store.EnsureDirs(); err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}
	server := app.NewServer(cfg, db, store)
	router := app.NewRouter(cfg, logger, server)
	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: router}
	go func() {
		logger.Info("server started", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
