package middlerware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/C0deNe0/go-boiler/internal/errs"
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	server *server.Server
}

func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{
		server: s,
	}
}

func (auth *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return echo.WrapMiddleware(clerkhttp.WithHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)

			response := map[string]string{
				"code":     "UNAUTHORIZED",
				"message":  "Unauthorized",
				"override": "false",
				"status":   "401",
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				auth.server.Logger.Error().Err(err).Str("function", "RequiredAuth").Dur("duration", time.Since(start)).Msg("failed to write JSON response")
			} else {
				auth.server.Logger.Error().Err(err).Str("function", "RequiredAuth").Dur("duration", time.Since(start)).Msg("could not get session claims from context")
			}
		}))))(func(c echo.Context) error {
		start := time.Now()
		claims, ok := clerk.SessionClaimsFromContext(c.Request().Context())

		if !ok {
			auth.server.Logger.Error().Str("function", "RequiredAuth").Str("required_id", GetRequestID(c)).Dur("duration", time.Since(start)).Msg("could not get session claims from context")

			return errs.NewUnauthorizedError("Unauthorized", false)

		}

		c.Set("user_id", claims.Subject)
		c.Set("user_role", claims.ActiveOrganizationRole)
		c.Set("permission", claims.Claims.ActiveOrganizationPermissions)

		auth.server.Logger.Info().Str("function", "RequireAuth").Str("user_id", claims.Subject).Str("request_id", GetRequestID(c)).Dur("duration", time.Since(start)).Msg("User authenticated successfully")
		return next(c)
	})
}
