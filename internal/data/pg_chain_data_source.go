package data

import (
	"context"
	"errors"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/grassrootseconomics/ethutils"
	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	PgChainDataOpts struct {
		Logg    *slog.Logger
		DSN     string
		Queries *PgQueries
	}

	PgChainData struct {
		logg    *slog.Logger
		db      *pgxpool.Pool
		queries *PgQueries
	}
)

func NewPgChainDataSource(o PgChainDataOpts) (*PgChainData, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	return &PgChainData{
		logg:    o.Logg,
		db:      dbPool,
		queries: o.Queries,
	}, nil
}

func (pg *PgChainData) Last10Tx(ctx context.Context, publicAddress string) ([]*api.Last10TxResponse, error) {
	var last10Tx []*api.Last10TxResponse

	if err := pgxscan.Select(ctx, pg.db, &last10Tx, pg.queries.Last10Tx, publicAddress); err != nil {
		return nil, err
	}

	return last10Tx, nil
}

func (pg *PgChainData) TokenHoldings(ctx context.Context, publicAddress string) ([]*api.TokenHoldings, error) {
	var tokenHoldings []*api.TokenHoldings

	if err := pgxscan.Select(ctx, pg.db, &tokenHoldings, pg.queries.TokenHoldings, publicAddress); err != nil {
		return nil, err
	}

	return tokenHoldings, nil
}

func (pg *PgChainData) ResolveAlias(ctx context.Context, alias string) (*api.AliasAddress, error) {
	// TODO: Implement graph resolver via fdw
	return &api.AliasAddress{
		Address: ethutils.ZeroAddress.Hex(),
	}, nil
}

func (pg *PgChainData) TokenDetails(ctx context.Context, tokenAddress string) (*api.TokenDetails, error) {
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

func (pg *PgChainData) PoolDetails(ctx context.Context, poolAddress string) (*api.PoolDetails, error) {
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

func (pg *PgChainData) PoolReverseDetails(ctx context.Context, poolSymbol string) (*api.PoolDetails, error) {
	row, err := pg.db.Query(ctx, pg.queries.PoolReverseDetails, poolSymbol)
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
