package data

type PgQueries struct {
	Last10Tx           string `query:"last-10-tx"`
	TokenHoldings      string `query:"token-holdings"`
	TokenDetails       string `query:"token-details"`
	PoolDetails        string `query:"pool-details"`
	PoolReverseDetails string `query:"pool-reverse-details"`
	TopPools           string `query:"top-active-pools"`
	Stables            string `query:"stables"`
}
