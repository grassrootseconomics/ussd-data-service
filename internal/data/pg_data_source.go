package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/goyesql/v2"
)

type (
	Queries struct {
		Last10Tx      string `query:"last-10-tx"`
		TokenHoldings string `query:"token-holdings"`
		TokenDetails  string `query:"token-details"`
		PoolDetails   string `query:"pool-details"`
	}

	PgOpts struct {
		Logg              *slog.Logger
		DSN               string
		QueriesFolderPath string
	}

	Pg struct {
		logg    *slog.Logger
		db      *pgxpool.Pool
		queries *Queries
	}
)

func NewPgStore(o PgOpts) (*Pg, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	queries, err := loadQueries(o.QueriesFolderPath)
	if err != nil {
		return nil, err
	}

	return &Pg{
		logg:    o.Logg,
		db:      dbPool,
		queries: queries,
	}, nil
}

func loadQueries(queriesPath string) (*Queries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &Queries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}

func (pg *Pg) Last10Tx(ctx context.Context, publicAddress string) ([]*api.Last10TxResponse, error) {
	var last10Tx []*api.Last10TxResponse

	if err := pgxscan.Select(ctx, pg.db, &last10Tx, pg.queries.Last10Tx, publicAddress); err != nil {
		return nil, err
	}

	return last10Tx, nil
}

func (pg *Pg) TokenHoldings(ctx context.Context, publicAddress string) ([]*api.TokenHoldings, error) {
	var tokenHoldings []*api.TokenHoldings

	if err := pgxscan.Select(ctx, pg.db, &tokenHoldings, pg.queries.TokenHoldings, publicAddress); err != nil {
		return nil, err
	}

	return tokenHoldings, nil
}

func (pg *Pg) TokenDetails(ctx context.Context, tokenAddress string) (*api.TokenDetails, error) {
	row, err := pg.db.Query(ctx, pg.queries.TokenDetails, tokenAddress)
	if err != nil {
		return nil, err
	}

	var tokenDetails api.TokenDetails
	if err := pgxscan.ScanOne(&tokenDetails, row); errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &tokenDetails, nil
}

func (pg *Pg) PoolDetails(ctx context.Context, poolAddress string) (*api.PoolDetails, error) {
	row, err := pg.db.Query(ctx, pg.queries.PoolDetails, poolAddress)
	if err != nil {
		return nil, err
	}

	var poolDetails api.PoolDetails
	if err := pgxscan.ScanOne(&poolDetails, row); errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &poolDetails, nil
}
