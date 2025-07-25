package api

import "time"

type (
	OKResponse struct {
		Ok          bool           `json:"ok"`
		Description string         `json:"description"`
		Result      map[string]any `json:"result"`
	}

	ErrResponse struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}

	Last10TxResponse struct {
		Sender          string    `json:"sender" db:"sender"`
		Recipient       string    `json:"recipient" db:"recipient"`
		TransferValue   string    `json:"transferValue" db:"transfer_value"`
		ContractAddress string    `json:"contractAddress" db:"contract_address"`
		TxHash          string    `json:"txHash" db:"tx_hash"`
		DateBlock       time.Time `json:"dateBlock" db:"date_block"`
		TokenSymbol     string    `json:"tokenSymbol" db:"token_symbol"`
		TokenDecimals   string    `json:"tokenDecimals" db:"token_decimals"`
		Success         bool      `json:"success" db:"success"`
	}

	TokenHoldings struct {
		TokenAddress  string `json:"tokenAddress" db:"contract_address"`
		TokenSymbol   string `json:"tokenSymbol" db:"token_symbol"`
		TokenDecimals string `json:"tokenDecimals" db:"token_decimals"`
		Balance       string `json:"balance"`
	}

	TokenDetails struct {
		TokenAddress  string `json:"tokenAddress" db:"token_address"`
		TokenSymbol   string `json:"tokenSymbol" db:"token_symbol"`
		TokenDecimals uint8  `json:"tokenDecimals" db:"token_decimals"`
		SinkAddress   string `json:"sinkAddress" db:"sink_address"`
		TokenName     string `json:"tokenName" db:"token_name"`
		CommodityName string `json:"tokenCommodity" db:"commodity_name"`
		Location      string `json:"tokenLocation" db:"location_name"`
	}

	PoolDetails struct {
		PoolName            string `json:"poolName" db:"pool_name"`
		PoolSymbol          string `json:"poolSymbol" db:"pool_symbol"`
		PoolContractAdrress string `json:"poolContractAddress" db:"contract_address"`
		LimiterAddress      string `json:"limiterAddress" db:"token_limiter_address"`
		VoucherRegistry     string `json:"voucherRegistry" db:"token_registry_address"`
	}

	AliasAddress struct {
		Address string `json:"address" db:"blockchain_address"`
	}

	TokenSwapRates struct {
		InRate        uint64 `json:"inRate" db:"in_rate"`
		OutRate       uint64 `json:"outRate" db:"out_rate"`
		InDecimals    uint8  `json:"inDecimals" db:"in_decimals"`
		OutDecimals   uint8  `json:"outDecimals" db:"out_decimals"`
		InTokenLimit  string `json:"inTokenLimit" db:"in_token_limit"`
		OutTokenLimit string `json:"outTokenLimit" db:"out_token_limit"`
	}
)
