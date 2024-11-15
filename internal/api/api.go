package api

import (
	"context"
	"crypto"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/grassrootseconomics/ussd-data-service/internal/data"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

type (
	APIOpts struct {
		VerifyingKey    crypto.PublicKey
		EnableMetrics   bool
		ListenAddress   string
		Logg            *slog.Logger
		PgDataSource    *data.Pg
		ChainDataSource *data.Chain
	}

	API struct {
		validator       httputil.ValidatorProvider
		verifyingKey    crypto.PublicKey
		router          *bunrouter.Router
		server          *http.Server
		logg            *slog.Logger
		pgDataSource    *data.Pg
		chainDataSource *data.Chain
	}
)

const (
	apiVersion = "/api/v1"
	slaTimeout = 10 * time.Second
)

func New(o APIOpts) *API {
	api := &API{
		validator:       httputil.NewValidator(""),
		verifyingKey:    o.VerifyingKey,
		logg:            o.Logg,
		pgDataSource:    o.PgDataSource,
		chainDataSource: o.ChainDataSource,
		router: bunrouter.New(
			bunrouter.WithNotFoundHandler(notFoundHandler),
			bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
		),
	}

	if o.EnableMetrics {
		api.router.GET("/metrics", metricsHandler)
	}

	api.router.WithGroup(apiVersion, func(g *bunrouter.Group) {
		if os.Getenv("DEV") != "" {
			g = g.Use(reqlog.NewMiddleware())
		}

		g = g.Use(api.authMiddleware)

		g.GET("/transfers/last10/:address", api.last10TxHandler)
		g.GET("/holdings/:address", api.tokenHoldingsHandler)
		g.GET("/token/:address", api.tokenDetailsHandler)
		g.GET("/pool/:address", api.poolDetailsHandler)
	})

	api.server = &http.Server{
		Addr:    o.ListenAddress,
		Handler: api.router,
	}

	return api
}

func (a *API) Start() error {
	a.logg.Info("API server starting", "address", a.server.Addr)
	if err := a.server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
