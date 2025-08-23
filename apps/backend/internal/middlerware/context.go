package middlerware

import (
	"context"

	"github.com/C0deNe0/go-boiler/internal/logger"
	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"github.com/newrelic/go-agent/v3/newrelic"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

type ContextEnhancer struct {
	server *server.Server
}

func NewContextEnhancer(s *server.Server) *ContextEnhancer {
	return &ContextEnhancer{
		server: s,
	}
}

func (ce *ContextEnhancer) EnhanceContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := GetRequestID(c)

			contextLogger := ce.server.Logger.With().
				Str("request_id", requestID).
				Str("method", c.Request().Method).
				Str("path", c.Path()).Str("ip", c.RealIP()).Logger()

			//add trace context if available

			if txn := newrelic.FromContext(c.Request().Context()); txn != nil {
				contextLogger = logger.WithTraceContext(contextLogger, txn)
			}

			if userId := ce.extractUserID(c); userId != "" {
				contextLogger = contextLogger.With().Str("user_id", userId).Logger()
			}

			if userRole := ce.extractUserRole(c); userRole != "" {
				contextLogger = contextLogger.With().Str("user_id", userRole).Logger()
			}

			c.Set(LoggerKey, &contextLogger)

			ctx := context.WithValue(c.Request().Context(), LoggerKey, &contextLogger)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func (ce *ContextEnhancer) extractUserID(c echo.Context) string {
	if userID, ok := c.Get("user_id").(string); ok && userID != "" {
		return userID
	}

	return ""
}

func (ce *ContextEnhancer) extractUserRole(c echo.Context) string {
	if userRole, ok := c.Get("user_role").(string); ok && userRole != "" {
		return userRole
	}
	return ""
}

func GetUserID(c echo.Context) string {
	if userID, ok := c.Get(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

func GetLogger(c echo.Context) *zerolog.Logger {
	if logger, ok := c.Get(LoggerKey).(*zerolog.Logger); ok {
		return logger
	}
	// Fallback to a basic logger if not found
	logger := zerolog.Nop()
	return &logger
}
