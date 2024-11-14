package xapi

import (
	"net/http"

	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

func notFoundHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusNotFound, map[string]any{
		"ok":          false,
		"description": "Not found",
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, _ bunrouter.Request) error {
	return httputil.JSON(w, http.StatusMethodNotAllowed, map[string]any{
		"ok":          false,
		"description": "Method not allowed",
	})
}
