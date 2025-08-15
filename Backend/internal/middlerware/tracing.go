package middlerware

import (
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type TracingMiddleware struct {
	server *server.Server
	nrApp  *newrelic.Application
}

func NewTracingMiddleware(s *server.Server, nrApp *newrelic.Application) *TracingMiddleware {
	return &TracingMiddleware{
		server: s,
		nrApp:  nrApp,
	}
}

func (tm *TracingMiddleware) NewRelicMiddleware() echo.MiddlewareFunc {
	if tm.nrApp == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}
	return nrecho.Middleware(tm.nrApp)
}

func (tm *TracingMiddleware) EnhanceTracing() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			txn := newrelic.FromContext(c.Request().Context())
			if txn == nil {
				return next(c)
			}

			txn.AddAttribute("service.name", tm.server.Config.Observeability.ServiceName)
			txn.AddAttribute("service.environment", tm.server.Config.Observeability.Environment)
			txn.AddAttribute("http.reap_id", c.RealIP())
			txn.AddAttribute("http.user_agent", c.Request().UserAgent())

			//add request id if available
			if requestID := GetRequestID(c); requestID != "" {
				txn.AddAttribute("request.id", requestID)
			}

			if userID := c.Get("user_id"); userID != nil {
				if userIDStr, ok := userID.(string); ok {
					txn.AddAttribute("user.id", userIDStr)
				}
			}
			//executing next handler
			err := next(c)
			if err != nil {
				txn.NoticeError(nrpkgerrors.Wrap(err))
			}

			//add response status
			txn.AddAttribute("http.status_code", c.Response().Status)

			return err
		}
	}
}
