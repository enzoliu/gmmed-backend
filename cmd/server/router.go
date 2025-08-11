package main

import (
	"breast-implant-warranty-system/core/migration"
	"breast-implant-warranty-system/core/router"
	"breast-implant-warranty-system/core/singleton"
	"breast-implant-warranty-system/internal/middleware"
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

func SetupRouter(ctx context.Context, cfg *Config) (*router.ExtendedRouter, error) {
	e := echo.New()

	registerHealthRoutes(e)

	routerConfig := router.RouterConfig{
		ServiceName:   "main",
		EnableCORS:    true,
		CORSList:      cfg.CORSAllowedOrigins(),
		EnableRecover: true,
	}
	sub, err := router.SetupSubRouter(e, routerConfig)
	if err != nil {
		return nil, err
	}

	e.Use(middleware.AuditLog())

	extendedRouter := &router.ExtendedRouter{Echo: e, RouterItf: sub}

	singletonGroup := new(singleton.Group)

	// migrate
	if err := migration.Migrate(ctx, singletonGroup, cfg); err != nil {
		return nil, errors.Wrap(err, "failed to migrate database")
	}

	// register routes
	router.RegisterGMMedRoutes(ctx, extendedRouter, singletonGroup, cfg)

	return extendedRouter, nil
}

func registerHealthRoutes(r router.RouterItf) {
	healthHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	}

	g := r.Group("")
	g.GET("/", healthHandler)
	g.GET("/health", healthHandler)
}
