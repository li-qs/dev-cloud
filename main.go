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
	"devcloud/web/handler"
	"devcloud/worker"

	"github.com/labstack/echo/v5"
)

func main() {
	initLogger()
	if err := run(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func initLogger() {
	slogOpts := &slog.HandlerOptions{}
	if os.Getenv("ENV") == "development" {
		slogOpts.Level = slog.LevelDebug
	} else {
		slogOpts.Level = slog.LevelInfo
	}
	slog.SetDefault(
		slog.New(slog.NewJSONHandler(os.Stdout, slogOpts)),
	)
}

func run() error {
	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()
	var cfg config.Config
	if err := config.LoadFile(configFile, &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	psql, sqlDB, err := storage.NewPostgres(&cfg)
	if err != nil {
		return fmt.Errorf("open postgresql: %w", err)
	}
	defer psql.Close()

	if os.Getenv("ENV") == "development" {
		if err := psql.Schema.Create(ctx); err != nil {
			return fmt.Errorf("create database schema: %w", err)
		}
	}

	rdb, err := storage.NewRedis(&cfg)
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer rdb.Close()

	repo := repo.New(psql)
	store := store.New(rdb)

	providers, dockerProvider, err := initProviders(&cfg)
	if err != nil {
		return fmt.Errorf("init providers: %w", err)
	}

	workerPool := worker.NewPool(ctx, cfg.WorkerPool, repo, providers)
	go workerPool.Start()
	defer workerPool.Stop()

	checks := []handler.HealthCheck{
		{Name: "postgres", Check: sqlDB.PingContext},
		{Name: "redis", Check: func(ctx context.Context) error { return rdb.Ping(ctx).Err() }},
		{Name: "docker", Check: dockerProvider.Ping},
	}

	webErr := make(chan error, 1)
	webServer := web.NewServer(&cfg, repo, store, checks)
	go func() {
		sc := echo.StartConfig{
			Address:    cfg.ServerAddr,
			HideBanner: true,
		}
		webErr <- sc.Start(ctx, webServer)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown", "cause", ctx.Err())
	case err := <-webErr:
		if err != nil {
			return fmt.Errorf("web server: %w", err)
		}
		slog.Info("web server stopped")
	}

	return nil
}

func initProviders(c *config.Config) (*provider.Registry, *docker.Docker, error) {
	registry := provider.NewRegistry()

	d, err := docker.NewProvider(
		c.Docker.Host,
		c.Docker.TLS.Enabled,
		c.Docker.TLS.CA,
		c.Docker.TLS.Cert,
		c.Docker.TLS.Key,
	)
	if err != nil {
		return nil, nil, err
	}
	registry.Register(
		resource.ProviderDocker,
		d,
	)

	return registry, d, nil
}
