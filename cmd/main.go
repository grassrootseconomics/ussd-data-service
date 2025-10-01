package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grassrootseconomics/ussd-data-service/internal/api"
	"github.com/grassrootseconomics/ussd-data-service/internal/data"
	"github.com/grassrootseconomics/ussd-data-service/internal/util"
	"github.com/knadh/goyesql/v2"
	"github.com/knadh/koanf/v2"
)

const defaultGracefulShutdownPeriod = time.Second * 20

var (
	build = "dev"

	confFlag    string
	queriesFlag string

	lo *slog.Logger
	ko *koanf.Koanf
)

func init() {
	flag.StringVar(&confFlag, "config", "config.toml", "Config file location")
	flag.StringVar(&queriesFlag, "queries", "queries.sql", "Queries file location")
	flag.Parse()

	lo = util.InitLogger()
	ko = util.InitConfig(lo, confFlag)

	lo.Info("starting ussd data service", "build", build)
}

func main() {
	var wg sync.WaitGroup
	ctx, stop := notifyShutdown()

	pgQueries, err := loadQueries(queriesFlag)
	if err != nil {
		lo.Error("could not load queries", "error", err)
		os.Exit(1)
	}

	pgChainDataStore, err := data.NewPgChainDataSource(data.PgChainDataOpts{
		Logg:    lo,
		DSN:     ko.MustString("postgres.federation_dsn"),
		Queries: pgQueries,
	})
	if err != nil {
		lo.Error("could not initialize postgres store", "error", err)
		os.Exit(1)
	}

	chainData := data.NewChainProvider(data.ChainOpts{
		ChainID:         ko.MustInt64("chain.id"),
		RPCEndpoint:     ko.MustString("chain.rpc_endpoint"),
		BalancesScanner: ko.MustString("chain.balances_scanner"),
		Logg:            lo,
	})

	publicKey, err := util.LoadSigningKey(ko.MustString("api.public_key"))
	if err != nil {
		lo.Error("could not load private key", "error", err)
		os.Exit(1)
	}

	apiServer := api.New(api.APIOpts{
		VerifyingKey:    publicKey,
		EnableMetrics:   ko.Bool("metrics.enable"),
		ListenAddress:   ko.MustString("api.address"),
		PgDataSource:    pgChainDataStore,
		ChainDataSource: chainData,
		Logg:            lo,
	})

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Start(); err != nil {
			lo.Error("failed to start HTTP server", "err", fmt.Sprintf("%T", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	lo.Info("shutdown signal received")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultGracefulShutdownPeriod)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiServer.Stop(shutdownCtx); err != nil {
			lo.Error("failed to stop HTTP server", "err", fmt.Sprintf("%T", err))
		}
	}()

	go func() {
		wg.Wait()
		stop()
		cancel()
		os.Exit(0)
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		stop()
		cancel()
		lo.Error("graceful shutdown period exceeded, forcefully shutting down")
	}
	os.Exit(1)
}

func notifyShutdown() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
}

func loadQueries(queriesPath string) (*data.PgQueries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &data.PgQueries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}
