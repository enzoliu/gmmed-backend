package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"breast-implant-warranty-system/pkg/cfgloader"
	"breast-implant-warranty-system/pkg/graceful"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := runMain(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func runMain() error {
	shutdowner, ctx := graceful.StartShutdowner(context.Background())
	defer shutdowner.CloseAndWait()

	// load config
	cfg, err := cfgloader.Load[Config](ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}
	// 設定日誌
	setupLogger(cfg)

	// setup router
	router, err := SetupRouter(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "failed to setup router")
	}

	// run server
	shutdowner.DeferShutdownFunc(func(ctx context.Context) {
		slog.InfoContext(ctx, "shutting down server")
		if err := router.Shutdown(ctx); err != nil {
			slog.ErrorContext(ctx, "failed to shutdown server", "err", err.Error())
		}
		slog.InfoContext(ctx, "server exited")
	})

	if err := router.StartAndListen(ctx, fmt.Sprintf(":%s", cfg.PORT)); err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "failed to start server")
	}
	return nil
}

func setupLogger(cfg *Config) {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(cfg.LOG_LEVEL)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if cfg.DEBUG {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
