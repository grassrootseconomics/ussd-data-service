package api

import (
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

	if err := a.chainDataSource.MergeTokenBalances(req.Context(), tokenHoldings, r.Address); err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Token holdings with current balances",
		Result: map[string]any{
			"holdings": tokenHoldings,
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

	poolDetails, err := a.chainDataSource.PoolDetails(req.Context(), u.PoolAddress)
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

	tokenHoldings, err := a.pgDataSource.TokenHoldings(req.Context(), u.UserAddress)
	if err != nil {
		return err
	}

	filtered, err := a.chainDataSource.TokensExistsInIndex(req.Context(), poolDetails.VoucherRegistry, tokenHoldings)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap from list",
		Result: map[string]any{
			"filtered": filtered,
		},
	})
}

func (a *API) poolSwapFromCheck(w http.ResponseWriter, req bunrouter.Request) error {
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

	poolDetails, err := a.chainDataSource.PoolDetails(req.Context(), u.PoolAddress)
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

	exists, err := a.chainDataSource.TokenExistsInIndex(req.Context(), poolDetails.VoucherRegistry, u.UserAddress)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap from check",
		Result: map[string]any{
			"canSwapFrom": exists,
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

	poolDetails, err := a.chainDataSource.PoolDetails(req.Context(), u.Address)
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

	if isStablesQueryOnly {
		stables, err := a.pgDataSource.Stables(req.Context())
		if err != nil {
			return err
		}

		filtered, err := a.chainDataSource.TokensExistsInIndex(req.Context(), poolDetails.VoucherRegistry, stables)
		if err != nil {
			return err
		}

		return httputil.JSON(w, http.StatusOK, api.OKResponse{
			Ok:          true,
			Description: "Swap to list",
			Result: map[string]any{
				"filtered": filtered,
			},
		})
	}

	allTokens, err := a.chainDataSource.AllTokensInIndex(req.Context(), poolDetails.VoucherRegistry)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Swap to list",
		Result: map[string]any{
			"filtered": allTokens,
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

	if err := a.validator.Validate(u); err != nil {
		return httputil.JSON(w, http.StatusBadRequest, api.ErrResponse{
			Ok:          false,
			Description: "Address validation failed",
		})
	}

	poolDetails, err := a.chainDataSource.PoolDetails(req.Context(), u.PoolAddress)
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

	maxLimit, err := a.chainDataSource.MaxLimit(
		req.Context(),
		u.UserAddress,
		u.PoolAddress,
		poolDetails.LimiterAddress,
		u.FromToken,
		u.ToToken,
	)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "From token max limit",
		Result: map[string]any{
			"max": maxLimit.String(),
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
