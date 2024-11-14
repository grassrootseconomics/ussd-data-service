package xapi

import (
	"net/http"

	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type PublicAddressParam struct {
	Address string `validate:"required,eth_addr_checksum"`
}

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
		// TODO: Handle error
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

	tokenDetails, err := a.chainDataSource.TokenDetails(req.Context(), r.Address)
	if err != nil {
		return err
	}

	return httputil.JSON(w, http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Token details",
		Result: map[string]any{
			"tokenDetails": tokenDetails,
		},
	})
}
