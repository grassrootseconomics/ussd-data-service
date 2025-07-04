package api

import (
	"math/big"
	"net/http"

	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type (
	PublicAddressParam struct {
		Address string `validate:"required,eth_addr_checksum"`
	}

	SymbolParam struct {
		Symbol string `validate:"required"`
	}

	PoolVoucherList struct {
		UserAddress string `validate:"required,eth_addr_checksum"`
		PoolAddress string `validate:"required,eth_addr_checksum"`
	}

	TokenList struct {
		TokenAddress string `validate:"required,eth_addr_checksum"`
		PoolAddress  string `validate:"required,eth_addr_checksum"`
	}

	PoolLimits struct {
		PoolAddress string `validate:"required,eth_addr_checksum"`
		UserAddress string `validate:"required,eth_addr_checksum"`
		FromToken   string `validate:"required,eth_addr_checksum"`
		ToToken     string `validate:"required,eth_addr_checksum"`
	}

	AliasParam struct {
		// TODO: Add extra validations here
		Alias string `validate:"required"`
	}
)

func (a *API) last10TxHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := PublicAddressParam{
		Address: req.Param("address"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	last10Tx, err := a.pgDataSource.Last10Tx(req.Context(), r.Address)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Last 10 token transfers",
		Result: map[string]any{
			"transfers": last10Tx,
		},
	})
}

func (a *API) tokenHoldingsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := PublicAddressParam{
		Address: req.Param("address"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	tokenHoldings, err := a.pgDataSource.TokenHoldings(req.Context(), r.Address)
	if err != nil {
		return err
	}

	filteredHoldings, err := a.chainDataSource.MergeTokenBalances(req.Context(), tokenHoldings, r.Address)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Token holdings with current balances",
		Result: map[string]any{
			"holdings": filteredHoldings,
		},
	})
}

func (a *API) tokenDetailsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := PublicAddressParam{
		Address: req.Param("address"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}
	tokenDetails, err := a.pgDataSource.TokenDetails(req.Context(), r.Address)
	if err != nil {
		a.logg.Error("Failed to get token details", "error", err)
		return err
	}

	if tokenDetails == nil {
		tokenDetails, err = a.chainDataSource.TokenDetails(req.Context(), r.Address)
		if err != nil {
			return err
		}
	}

	// TODO: Implement graph resolver via fdw
	tokenDetails.CommodityName = "Farming"
	tokenDetails.Location = "Nairobi, Kenya"

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Token details",
		Result: map[string]any{
			"tokenDetails": tokenDetails,
		},
	})
}

func (a *API) poolDetailsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := PublicAddressParam{
		Address: req.Param("address"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	poolDetails, err := a.pgDataSource.PoolDetails(req.Context(), r.Address)
	if err != nil {
		return err
	}

	if poolDetails == nil {
		poolDetails, err = a.chainDataSource.PoolDetails(req.Context(), r.Address)
		if err != nil {
			return err
		}
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Pool details",
		Result: map[string]any{
			"poolDetails": poolDetails,
		},
	})
}

func (a *API) poolReverseDetailsHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := SymbolParam{
		Symbol: req.Param("symbol"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	poolDetails, err := a.pgDataSource.PoolReverseDetails(req.Context(), r.Symbol)
	if err != nil {
		a.logg.Debug("Failed to get pool details", "error", err)
		return err
	}

	if poolDetails == nil {
		return httputil.JSON(w, http.StatusNotFound, api.ErrResponse{
			Ok:          false,
			Description: "Pool not found",
		})
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Pool details",
		Result: map[string]any{
			"poolDetails": poolDetails,
		},
	})
}

func (a *API) topPoolsHandlder(w http.ResponseWriter, req bunrouter.Request) error {
	topPools, err := a.pgDataSource.TopPools(req.Context())
	if err != nil {
		a.logg.Debug("Failed to get pool details", "error", err)
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Top 5 pools sorted by swaps",
		Result: map[string]any{
			"topPools": topPools,
		},
	})
}

func (a *API) poolSwapFromVouchersList(w http.ResponseWriter, req bunrouter.Request) error {
	u := PoolVoucherList{
		UserAddress: req.Param("address"),
		PoolAddress: req.Param("pool"),
	}

	if err := a.validator.Validate(u); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	poolDetails, err := a.pgDataSource.PoolDetails(req.Context(), u.PoolAddress)
	if err != nil {
		a.logg.Debug("Failed to get pool details", "error", err)
		return err
	}

	if poolDetails == nil {
		return httputil.JSON(w, http.StatusNotFound, api.ErrResponse{
			Ok:          false,
			Description: "Pool not found",
		})
	}

	filtered, err := a.pgDataSource.PoolAllowedTokensForUser(req.Context(), u.UserAddress, u.PoolAddress)
	if err != nil {
		return err
	}

	filteredHoldings, err := a.chainDataSource.MergeTokenBalances(req.Context(), filtered, u.UserAddress)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap from list",
		Result: map[string]any{
			"filtered": filteredHoldings,
		},
	})
}

func (a *API) poolSwapFromCheck(w http.ResponseWriter, req bunrouter.Request) error {
	u := TokenList{
		TokenAddress: req.Param("address"),
		PoolAddress:  req.Param("pool"),
	}

	if err := a.validator.Validate(u); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	isAllowed, err := a.pgDataSource.PoolTokenAllowed(req.Context(), u.PoolAddress, u.TokenAddress)
	if err != nil {
		a.logg.Debug("Failed to check if token is allowed in pool", "error", err)
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap from check",
		Result: map[string]any{
			"canSwapFrom": isAllowed,
		},
	})
}

func (a *API) poolSwapToVouchersList(w http.ResponseWriter, req bunrouter.Request) error {
	u := PublicAddressParam{
		Address: req.Param("pool"),
	}
	isStablesQueryOnly := req.URL.Query().Get("stables") == "true"

	if err := a.validator.Validate(u); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	if isStablesQueryOnly {
		stables, err := a.pgDataSource.PoolAllowedStables(req.Context(), u.Address)
		if err != nil {
			return err
		}

		return httputil.JSON(w, http.StatusOK, api.OKResponse{
			Ok:          true,
			Description: "Swap to list (stables only)",
			Result: map[string]any{
				"filtered": stables,
			},
		})
	}

	allTokens, err := a.pgDataSource.PoolAllowedTokens(req.Context(), u.Address)
	if err != nil {
		return err
	}

	filteredHoldings, err := a.chainDataSource.MergeTokenBalances(req.Context(), allTokens, u.Address)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap to list (all tokens)",
		Result: map[string]any{
			"filtered": filteredHoldings,
		},
	})
}

func (a *API) poolMaxLimit(w http.ResponseWriter, req bunrouter.Request) error {
	u := PoolLimits{
		PoolAddress: req.Param("pool"),
		UserAddress: req.Param("address"),
		FromToken:   req.Param("from"),
		ToToken:     req.Param("to"),
	}
	a.logg.Debug("Pool max limit request", "pool", u.PoolAddress, "user", u.UserAddress, "from", u.FromToken, "to", u.ToToken)

	if err := a.validator.Validate(u); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	swapRates, err := a.pgDataSource.PoolTokenSwapRates(req.Context(), u.PoolAddress, u.FromToken, u.ToToken)
	if err != nil {
		a.logg.Debug("Failed to get token swap rates", "error", err)
		return err
	}

	if swapRates == nil {
		return httputil.JSON(w, http.StatusNotFound, api.ErrResponse{
			Ok:          false,
			Description: "Token swap rates not found for this pool",
		})
	}
	a.logg.Debug("Swap rates found", "inRate", swapRates.InRate, "outRate", swapRates.OutRate,
		"inDecimals", swapRates.InDecimals, "outDecimals", swapRates.OutDecimals,
		"inTokenLimit", swapRates.InTokenLimit, "outTokenLimit", swapRates.OutTokenLimit)

	// Get user balance and pool balance from chain
	userInBalance, poolInBalance, poolOutBalance, err := a.chainDataSource.GetSwapBalances(
		req.Context(),
		u.UserAddress,
		u.PoolAddress,
		u.FromToken,
		u.ToToken,
	)
	if err != nil {
		return err
	}

	// Convert the token limit from database string to *big.Int
	inTokenLimit := new(big.Int)
	if _, ok := inTokenLimit.SetString(swapRates.InTokenLimit, 10); !ok {
		return httputil.JSON(w, http.StatusInternalServerError, api.ErrResponse{
			Ok:          false,
			Description: "Invalid token limit format",
		})
	}

	outTokenLimit := new(big.Int)
	if _, ok := outTokenLimit.SetString(swapRates.OutTokenLimit, 10); !ok {
		return httputil.JSON(w, http.StatusInternalServerError, api.ErrResponse{
			Ok:          false,
			Description: "Invalid token limit format",
		})
	}

	maxSwapInput := a.chainDataSource.MaxSwapInput(
		userInBalance,
		inTokenLimit,
		outTokenLimit,
		poolInBalance,
		poolOutBalance,
		swapRates.InRate,
		swapRates.OutRate,
		swapRates.InDecimals,
		swapRates.OutDecimals,
	)

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "From token max swap input",
		Result: map[string]any{
			"max": maxSwapInput.String(),
		},
	})
}

func (a *API) aliasHandler(w http.ResponseWriter, req bunrouter.Request) error {
	r := AliasParam{
		Alias: req.Param("alias"),
	}

	if err := a.validator.Validate(r); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Alias validation failed",
		})
	}

	aliasAddress, err := a.pgDataSource.ResolveAlias(req.Context(), r.Alias)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Alias address",
		Result: map[string]any{
			"address": aliasAddress.Address,
		},
	})
}
