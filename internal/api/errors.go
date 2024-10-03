package api

import (
	"net/http"

	"github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/labstack/echo/v4"
)

func newBadRequestError(message ...interface{}) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, message...)
}

func (a *API) customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if hErr, ok := err.(*echo.HTTPError); ok {
		var errorMsg string

		if m, ok := hErr.Message.(error); ok {
			errorMsg = m.Error()
		} else if m, ok := hErr.Message.(string); ok {
			errorMsg = m
		}

		c.JSON(hErr.Code, api.ErrResponse{
			Ok:          false,
			Description: errorMsg,
		})
		return
	}

	a.logg.Error("api: echo error", "path", c.Path(), "err", err)
	c.JSON(http.StatusInternalServerError, api.ErrResponse{
		Ok:          false,
		Description: "Internal server error",
		ErrCode:     "E01",
	})
}

func handleBindError(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, api.ErrResponse{
		Ok:          false,
		ErrCode:     "E02",
		Description: "Invalid or malformed JSON structure",
	})
}

func handleValidateError(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, api.ErrResponse{
		Ok:          false,
		ErrCode:     "E04",
		Description: "Validation failed on one or more fields",
	})
}
