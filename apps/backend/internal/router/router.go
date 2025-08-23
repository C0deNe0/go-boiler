package router

import (
	"net/http"

	"github.com/C0deNe0/go-boiler/internal/handler"
	"github.com/C0deNe0/go-boiler/internal/middlerware"
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/C0deNe0/go-boiler/internal/service"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func NewRouter(s *server.Server, h *handler.Handlers, services *service.Services) *echo.Echo {
	middlewares := middlerware.NewMiddlewares(s)
	router := echo.New()

	router.HTTPErrorHandler = middlewares.Global.GlobalErrorHandler

	//Global Middlewares
	router.Use(
		echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
			Store: echoMiddleware.NewRateLimiterMemoryStore(rate.Limit(20)),
			DenyHandler: func(c echo.Context, identifier string, err error) error {
				//recording rate limit hit metrices
				if rateLimitMiddleware := middlewares.RateLimit; rateLimitMiddleware != nil {
					rateLimitMiddleware.RecordRateLimitHit(c.Path())
				}

				s.Logger.Warn().Str("request_id", middlerware.GetRequestID(c)).
					Str("identifier", identifier).
					Str("path", c.Path()).
					Str("method", c.Request().Method).
					Str("ip", c.RealIP()).
					Msg("rate limit exceeded")

				return echo.NewHTTPError(http.StatusBadRequest, "Rate limit exceeded")
			},
		}),
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		middlerware.RequestID(),
		middlewares.Tracing.NewRelicMiddleware(),
		middlewares.Tracing.EnhanceTracing(),
		middlewares.ContextEnhancer.EnhanceContext(),
		middlewares.Global.RequestLogger(),
		middlewares.Global.Recover(),
	)
	//registering systemRoutes
	registerSystemRoutes(router, h)

	//registered versioned routes
	router.Group("/api/v1")

	return router
}
