package handler

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed swagger_ui.html
var swaggerUIHTML string

//go:embed swagger.yaml
var swaggerYAML []byte

// SwaggerHandler serves Swagger UI and spec
type SwaggerHandler struct{}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// ServeUI serves the Swagger UI HTML page
func (h *SwaggerHandler) ServeUI(c echo.Context) error {
	return c.HTML(http.StatusOK, swaggerUIHTML)
}

// ServeSpec serves the OpenAPI specification
func (h *SwaggerHandler) ServeSpec(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml", swaggerYAML)
}
