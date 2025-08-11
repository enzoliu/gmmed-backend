package router

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/sync/errgroup"
)

const (
	EXTERNAL_API_RETRY = 5
)

type RouterItf interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(middleware ...echo.MiddlewareFunc)
	Group(prefix string, m ...echo.MiddlewareFunc) *echo.Group
}

type ExtendedRouter struct {
	RouterItf
	Echo *echo.Echo
}

func (s *ExtendedRouter) StartAndListen(ctx context.Context, addr string) error {
	wg, _ := errgroup.WithContext(ctx)

	wg.Go(func() error {
		return s.Echo.Start(addr)
	})

	return wg.Wait()
}

func (s *ExtendedRouter) Shutdown(ctx context.Context) error {
	return s.Echo.Shutdown(ctx)
}

type ExtendedRouterItf interface {
	RouterItf
	StartAndListen(ctx context.Context, addr string) error
	Shutdown(ctx context.Context) error
}

type RouterConfig struct {
	ServiceName   string
	EnableCORS    bool
	CORSList      string
	EnableRecover bool
}

func SetupSubRouter(rootRouter RouterItf, cfg RouterConfig) (RouterItf, error) {
	router := rootRouter.Group("")

	// use request id middleware
	router.Use(middleware.RequestID())

	// use cors middleware
	if cfg.EnableCORS {
		if cfg.CORSList == "" {
			cfg.CORSList = "http://localhost:4200,http://127.0.0.1:4200"
		}
		router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     strings.Split(cfg.CORSList, ","),
			AllowCredentials: true,
		}))
	}

	// use secure middleware
	// reference: https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html#basic-non-strict-csp-policy
	secureConfig := middleware.DefaultSecureConfig
	secureConfig.ContentSecurityPolicy = "default-src 'self'; frame-ancestors 'self'; form-action 'self';"
	router.Use(middleware.SecureWithConfig(secureConfig))

	return router, nil
}
