package data

import (
	"context"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/ethutils"
	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

type (
	ChainOpts struct {
		ChainID     int64
		RPCEndpoint string
		Logg        *slog.Logger
	}

	Chain struct {
		logg  *slog.Logger
		chain *ethutils.Provider
	}
)

func NewChainProvider(o ChainOpts) *Chain {
	return &Chain{
		logg:  o.Logg,
		chain: ethutils.NewProvider(o.RPCEndpoint, o.ChainID),
	}
}

func (c *Chain) MergeTokenBalances(ctx context.Context, input []*api.TokenHoldings, ownerAddress string) error {
	addresses := make([]common.Address, len(input))
	for i, holding := range input {
		addresses[i] = common.HexToAddress(holding.ContractAddress)
	}

	tokenBalances, err := c.chain.TokensBalance(ctx, common.HexToAddress(ownerAddress), addresses)
	if err != nil {
		return nil
	}

	for _, holding := range input {
		contractAddress := common.HexToAddress(holding.ContractAddress)
		if balance, exists := tokenBalances[contractAddress]; exists {
			holding.Balance = balance.String()
		}
	}

	return nil
}
