package api

import (
	"net/http"

	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/labstack/echo/v4"
)

type PublicAddressParam struct {
	Address string `param:"address"  validate:"required,eth_addr_checksum"`
}

func (a *API) last10TxHandler(c echo.Context) error {
	req := PublicAddressParam{}

	if err := c.Bind(&req); err != nil {
		return handleBindError(c)
	}

	if err := c.Validate(req); err != nil {
		return handleValidateError(c)
	}

	last10Tx, err := a.pgDataStore.Last10Tx(c.Request().Context(), req.Address)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Last 10 token transfers",
		Result: map[string]any{
			"transfers": last10Tx,
		},
	})
}

func (a *API) tokenHoldingsHandler(c echo.Context) error {
	req := PublicAddressParam{}

	if err := c.Bind(&req); err != nil {
		return handleBindError(c)
	}

	if err := c.Validate(req); err != nil {
		return handleValidateError(c)
	}

	tokenHoldings, err := a.pgDataStore.TokenHoldings(c.Request().Context(), req.Address)
	if err != nil {
		return err
	}

	if err := a.chainData.MergeTokenBalances(c.Request().Context(), tokenHoldings, req.Address); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, api.OKResponse{
		Ok:          true,
		Description: "Token holdings with current balances",
		Result: map[string]any{
			"holdings": tokenHoldings,
		},
	})
}
