package api

import (
	"net/http"

	model "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func notFoundHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusNotFound, model.ErrResponse{
		Ok:          false,
		Description: "Not found",
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusMethodNotAllowed, model.ErrResponse{
		Ok:          false,
		Description: "Method not allowed",
	})
}
