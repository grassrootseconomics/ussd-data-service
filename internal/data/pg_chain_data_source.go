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

func (pg *PgChainData) TopPools(ctx context.Context) ([]*api.PoolDetails, error) {
	var topPools []*api.PoolDetails

	if err := pgxscan.Select(ctx, pg.db, &topPools, pg.queries.TopPools); err != nil {
		return nil, err
	}

	return topPools, nil
}

func (pg *PgChainData) PoolAllowedTokensForUser(ctx context.Context, userAddress, poolAddress string) ([]*api.TokenHoldings, error) {
	var tokenHoldings []*api.TokenHoldings

	if err := pgxscan.Select(ctx, pg.db, &tokenHoldings, pg.queries.PoolAllowedTokensForUser, userAddress, poolAddress); err != nil {
		return nil, err
	}

	return tokenHoldings, nil
}

func (pg *PgChainData) PoolTokenAllowed(ctx context.Context, poolAddress, tokenAddress string) (bool, error) {
	var result struct {
		IsAllowed bool `db:"is_allowed"`
	}

	row, err := pg.db.Query(ctx, pg.queries.PoolTokenAllowed, poolAddress, tokenAddress)
	if err != nil {
		return false, err
	}

	if err := pgxscan.ScanOne(&result, row); err != nil {
		return false, err
	}

	return result.IsAllowed, nil
}

func (pg *PgChainData) PoolAllowedTokens(ctx context.Context, poolAddress string) ([]*api.TokenHoldings, error) {
	var tokenHoldings []*api.TokenHoldings

	if err := pgxscan.Select(ctx, pg.db, &tokenHoldings, pg.queries.PoolAllowedTokens, poolAddress); err != nil {
		return nil, err
	}

	return tokenHoldings, nil
}

func (pg *PgChainData) PoolAllowedStables(ctx context.Context, poolAddress string) ([]*api.TokenHoldings, error) {
	var tokenHoldings []*api.TokenHoldings

	if err := pgxscan.Select(ctx, pg.db, &tokenHoldings, pg.queries.PoolAllowedStables, poolAddress); err != nil {
		return nil, err
	}

	return tokenHoldings, nil
}

func (pg *PgChainData) PoolTokenSwapRates(ctx context.Context, poolAddress, inTokenAddress, outTokenAddress string) (*api.TokenSwapRates, error) {
	row, err := pg.db.Query(ctx, pg.queries.PoolTokenSwapRates, poolAddress, inTokenAddress, outTokenAddress)
	if err != nil {
		return nil, err
	}

	var swapRates api.TokenSwapRates
	if err := pgxscan.ScanOne(&swapRates, row); errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &swapRates, nil
}

func (pg *PgChainData) PoolTokenLimit(ctx context.Context, poolAddress, tokenAddress string) (string, error) {
	var result struct {
		TokenLimit string `db:"token_limit"`
	}

	row, err := pg.db.Query(ctx, pg.queries.PoolTokenLimit, poolAddress, tokenAddress)
	if err != nil {
		return "", err
	}

	if err := pgxscan.ScanOne(&result, row); errors.Is(err, pgx.ErrNoRows) {
		return "0", nil
	} else if err != nil {
		return "", err
	}

	return result.TokenLimit, nil
}
