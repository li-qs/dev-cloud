package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"devcloud/config"
	"devcloud/ent/resource"
	"devcloud/provider"
	"devcloud/provider/docker"
	"devcloud/repo"
	"devcloud/storage"
	"devcloud/store"
	"devcloud/web"
	"devcloud/worker"

	"github.com/labstack/echo/v5"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	slogOpts := &slog.HandlerOptions{}
	if os.Getenv("ENV") == "development" {
		slogOpts.Level = slog.LevelDebug
	} else {
		slogOpts.Level = slog.LevelInfo
	}
	slog.SetDefault(
		slog.New(slog.NewJSONHandler(os.Stdout, slogOpts)),
	)

	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()
	var cfg config.Config
	if err := config.LoadFile(configFile, &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	psql, err := storage.NewPostgres(&cfg)
	if err != nil {
		return fmt.Errorf("open postgresql: %w", err)
	}
	defer psql.Close()

	if err := psql.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create database schema: %w", err)
	}

	rdb, err := storage.NewRedis(&cfg)
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer rdb.Close()

	repo := repo.New(psql)
	store := store.New(rdb)

	runErr := make(chan error, 1)

	webServer := web.NewServer(&cfg, repo, store)
	go func() {
		sc := echo.StartConfig{
			Address:    cfg.ServerAddr,
			HideBanner: true,
		}
		runErr <- sc.Start(ctx, webServer)
	}()

	registry := provider.NewRegistry()
	registry.Register(
		resource.ProviderDocker,
		docker.NewProvider(),
	)

	workerPool := worker.NewPool(ctx, 4, repo, registry)
	go workerPool.Start()
	defer workerPool.Stop()

	select {
	case <-ctx.Done():
		slog.Info("shutdown", "cause", ctx.Err())
	case err := <-runErr:
		if err != nil {
			return fmt.Errorf("web server: %w", err)
		}
		slog.Info("web server stopped")
	}

	return nil
}
