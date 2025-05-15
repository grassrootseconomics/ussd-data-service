package data

import (
	"context"
	"errors"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/ethutils"
	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/lmittmann/w3/w3types"
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
	exists                = w3.MustNewFunc("have(address)", "bool")
	balanceOf             = w3.MustNewFunc("balanceOf(address)", "uint256")
	limitOf               = w3.MustNewFunc("limitOf(address, address)", "uint256")
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

func (c *Chain) MaxLimit(ctx context.Context, initator string, poolAddress string, limiterAddress string, inToken string, outToken string) (*big.Int, error) {
	var (
		initiatorInTokenBalance *big.Int
		outTokenBalance         *big.Int
		inTokenLimit            *big.Int

		batchErr w3.CallErrors
	)

	if err := c.chain.Client.CallCtx(
		ctx,
		eth.CallFunc(common.HexToAddress(inToken), balanceOf, common.HexToAddress(initator)).Returns(&initiatorInTokenBalance),
		eth.CallFunc(common.HexToAddress(outToken), balanceOf, common.HexToAddress(poolAddress)).Returns(&outTokenBalance),
		eth.CallFunc(common.HexToAddress(limiterAddress), limitOf, common.HexToAddress(inToken), common.HexToAddress(poolAddress)).Returns(&inTokenLimit),
	); errors.As(err, &batchErr) {
		return nil, batchErr
	} else if err != nil {
		return nil, err
	}

	return min([]*big.Int{inTokenLimit, initiatorInTokenBalance, outTokenBalance}), nil
}

// This is very inefficent beacuse of round trips. But it is the only way to do it for now.
func (c *Chain) TokensExistsInIndex(ctx context.Context, index string, input []*api.TokenHoldings) ([]*api.TokenHoldings, error) {
	calls := make([]w3types.RPCCaller, len(input))
	resp := make([]bool, len(input))

	for i, holding := range input {
		calls[i] = eth.CallFunc(common.HexToAddress(index), exists, common.HexToAddress(holding.ContractAddress)).Returns(&resp[i])
	}

	var batchErr w3.CallErrors

	if err := c.chain.Client.CallCtx(
		ctx,
		calls...,
	); errors.As(err, &batchErr) {
		return nil, batchErr
	} else if err != nil {
		return nil, err
	}

	j := 0
	for i := 0; i < len(input); i++ {
		if resp[i] {
			input[j] = input[i]
			j++
		}
	}
	input = input[:j]
	return input, nil
}

func min(values []*big.Int) *big.Int {
	if len(values) == 0 {
		return nil
	}

	min := new(big.Int).Set(values[0])

	for _, val := range values[1:] {
		if val.Cmp(min) < 0 {
			min.Set(val)
		}
	}

	return min
}

func (c *Chain) TokenExistsInIndex(ctx context.Context, index string, tokenAddress string) (bool, error) {
	var existsResp bool

	err := c.chain.Client.CallCtx(
		ctx,
		eth.CallFunc(common.HexToAddress(index), exists, common.HexToAddress(tokenAddress)).Returns(&existsResp),
	)
	if err != nil {
		return false, err
	}
	return existsResp, nil
}
