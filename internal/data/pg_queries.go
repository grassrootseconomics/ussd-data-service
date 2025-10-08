package data

type PgQueries struct {
	Last10Tx                 string `query:"last-10-tx"`
	TokenHoldings            string `query:"token-holdings"`
	TokenDetails             string `query:"token-details"`
	PoolDetails              string `query:"pool-details"`
	PoolReverseDetails       string `query:"pool-reverse-details"`
	TopPools                 string `query:"top-active-pools"`
	PoolTokenAllowed         string `query:"pool-token-allowed"`
	PoolAllowedTokensForUser string `query:"pool-allowed-tokens-for-user"`
	PoolAllowedTokens        string `query:"pool-allowed-tokens"`
	PoolAllowedStables       string `query:"pool-allowed-stables"`
	PoolTokenSwapRates       string `query:"pool-token-swap-rates"`
	PoolTokenLimit           string `query:"pool-token-limit"`
}
