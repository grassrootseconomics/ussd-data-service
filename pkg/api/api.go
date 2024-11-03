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
		ErrCode     string `json:"errorCode"`
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
	}

	TokenHoldings struct {
		ContractAddress string `json:"contractAddress" db:"contract_address"`
		TokenSymbol     string `json:"tokenSymbol" db:"token_symbol"`
		TokenDecimals   string `json:"tokenDecimals" db:"token_decimals"`
		Balance         string `json:"balance"`
	}

	TokenDetails struct {
		TokenSymbol   string `json:"tokenSymbol" db:"token_symbol"`
		TokenDecimals string `json:"tokenDecimals" db:"token_decimals"`
		SinkAddress   string `json:"sinkAddress" db:"sink_address"`
		TokenName     string `json:"balance" db:"token_name"`
	}
)
