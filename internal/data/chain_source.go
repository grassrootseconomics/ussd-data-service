package data

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/ethutils"
	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
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

var (
	nameGetter            = w3.MustNewFunc("name()", "string")
	symbolGetter          = w3.MustNewFunc("symbol()", "string")
	decimalsGetter        = w3.MustNewFunc("decimals()", "uint8")
	sinkAddressGetter     = w3.MustNewFunc("sinkAddress()", "address")
	limiterAddressGetter  = w3.MustNewFunc("tokenLimiter()", "address")
	registryAddressGetter = w3.MustNewFunc("tokenRegistry()", "address")
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

func (c *Chain) TokenDetails(ctx context.Context, input string) (*api.TokenDetails, error) {
	contractAddress := w3.A(input)

	var (
		tokenName     string
		tokenSymbol   string
		tokenDecimals uint8
		sinkAddress   common.Address

		batchErr w3.CallErrors
	)

	if err := c.chain.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, nameGetter).Returns(&tokenName),
		eth.CallFunc(contractAddress, symbolGetter).Returns(&tokenSymbol),
		eth.CallFunc(contractAddress, decimalsGetter).Returns(&tokenDecimals),
	); errors.As(err, &batchErr) {
		return nil, batchErr
	} else if err != nil {
		return nil, err
	}

	if err := c.chain.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, sinkAddressGetter).Returns(&sinkAddress),
	); err != nil {
		// This will most likely revert if the contract does not have a sinkAddress
		// Instead of handling the error we just ignore it and set the value to 0
		sinkAddress = ethutils.ZeroAddress
	}

	return &api.TokenDetails{
		TokenName:     tokenName,
		TokenSymbol:   tokenSymbol,
		TokenDecimals: tokenDecimals,
		SinkAddress:   sinkAddress.Hex(),
	}, nil
}

func (c *Chain) PoolDetails(ctx context.Context, input string) (*api.PoolDetails, error) {
	contractAddress := w3.A(input)

	var (
		poolName             string
		poolSymbol           string
		tokenRegistryAddress common.Address
		limiterAddress       common.Address

		batchErr w3.CallErrors
	)

	if err := c.chain.Client.CallCtx(
		ctx,
		eth.CallFunc(contractAddress, nameGetter).Returns(&poolName),
		eth.CallFunc(contractAddress, symbolGetter).Returns(&poolSymbol),
		eth.CallFunc(contractAddress, registryAddressGetter).Returns(&tokenRegistryAddress),
		eth.CallFunc(contractAddress, limiterAddressGetter).Returns(&limiterAddress),
	); errors.As(err, &batchErr) {
		return nil, batchErr
	} else if err != nil {
		return nil, err
	}

	return &api.PoolDetails{
		PoolName:            poolName,
		PoolSymbol:          poolSymbol,
		PoolContractAdrress: input,
		LimiterAddress:      limiterAddress.Hex(),
		VoucherRegistry:     tokenRegistryAddress.Hex(),
	}, nil
}
