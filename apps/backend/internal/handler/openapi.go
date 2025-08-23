package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/C0deNe0/go-boiler/internal/server"
	"github.com/labstack/echo/v4"
)

type OpenAPIHandler struct {
	Handler
}

func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

func (h *OpenAPIHandler) ServeOpenAPIUI(c echo.Context) error {
	templateByte, err := os.ReadFile("static/openapi.html")

	c.Response().Header().Set("Cache-Control", "no-cache")
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI UI: %w", err)

	}

	templateString := string(templateByte)
	err = c.HTML(http.StatusOK, templateString)
	if err != nil {
		return fmt.Errorf("failed to write HTML responce: %w", err)
	}
	return nil
}
