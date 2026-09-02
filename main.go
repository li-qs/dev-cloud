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
	"devcloud/ent"
	"devcloud/web"

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	var cfg config.Config
	if err := config.LoadFile(configFile, &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := ent.Open("postgres", cfg.Postgres)
	if err != nil {
		return fmt.Errorf("open postgresql: %w", err)
	}
	defer client.Close()

	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create database schema: %w", err)
	}

	webServer := web.NewServer(&cfg, client)
	webErr := make(chan error, 1)
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
