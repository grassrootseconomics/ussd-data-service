package api

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	model "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/kamikazechaser/common/httputil"
	"github.com/uptrace/bunrouter"
)

type JWTCustomClaims struct {
	PublicKey string `json:"publicKey"`
	Service   bool   `json:"service"`
	jwt.RegisteredClaims
}

func (a *API) authMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		if h := req.Header.Get("Authorization"); h != "" {
			token, err := request.ParseFromRequest(req.Request, request.AuthorizationHeaderExtractor, func(t *jwt.Token) (interface{}, error) {
				if t.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
					return nil, jwt.ErrTokenUnverifiable
				}
				return a.verifyingKey, nil
			}, request.WithClaims(&JWTCustomClaims{}))

			if err != nil {
				a.logg.Error("JWT validation failed", "error", err)
				return httputil.JSON(w, http.StatusBadRequest, model.ErrResponse{
					Ok:          false,
					Description: "JWT validation failed",
				})
			}

			if !token.Valid {
				return httputil.JSON(w, http.StatusUnauthorized, model.ErrResponse{
					Ok:          false,
					Description: "Invalid token",
				})
			}

			if claims, ok := token.Claims.(JWTCustomClaims); ok {
				if !claims.Service {
					return httputil.JSON(w, http.StatusUnauthorized, model.ErrResponse{
						Ok:          false,
						Description: "Only service level keys allowed",
					})
				}
			}

			return next(w, req)
		} else {
			return httputil.JSON(w, http.StatusUnauthorized, model.ErrResponse{
				Ok:          false,
				Description: "Authorization token is required",
			})
		}
	}
}
