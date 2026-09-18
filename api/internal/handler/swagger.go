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

// SwaggerUIDistVersion pins the vendored swagger-ui-dist release.
//
// The three files under swaggerui/ used to be fetched from unpkg.com at page
// load, so /api/docs was blank on any install without egress to a CDN — a home
// server behind a restrictive firewall, an air-gapped LAN, or simply an unpkg
// outage — and the API's CSP had to whitelist a third-party origin in
// script-src, style-src and connect-src. They are embedded in the binary now.
//
// TO UPDATE: re-download all three from
// https://unpkg.com/swagger-ui-dist@<version>/<file> into swaggerui/, bump this
// constant, and bump the version recorded in THIRD-PARTY-LICENSES.md. Nothing
// else: swagger_ui.html's asset URLs are checked against this constant by
// swagger_ui_assets_test.go, which fails if the two drift.
const SwaggerUIDistVersion = "5.10.3"

// SwaggerAssetsRoute is the Echo pattern the vendored files are served under.
// The version segment is what makes the immutable Cache-Control in ServeAsset
// safe across an upgrade: a new release asks for a new URL instead of reusing a
// year-old cache entry.
const SwaggerAssetsRoute = "/api/docs/assets/:version/:file"

//go:embed swaggerui/swagger-ui.css
var swaggerUICSS []byte

//go:embed swaggerui/swagger-ui-bundle.js
var swaggerUIBundleJS []byte

//go:embed swaggerui/swagger-ui-standalone-preset.js
var swaggerUIStandalonePresetJS []byte

// swaggerUIAsset is one vendored file plus the Content-Type it must be served
// with — X-Content-Type-Options: nosniff is set on every API response, so a
// wrong or missing type means the browser refuses to execute the script.
type swaggerUIAsset struct {
	contentType string
	body        []byte
}

// swaggerUIAssets is the allow-list behind the :file route parameter. Nothing
// outside it is served, so the parameter can never name anything but these
// three embedded blobs.
var swaggerUIAssets = map[string]swaggerUIAsset{
	"swagger-ui.css":                  {"text/css; charset=utf-8", swaggerUICSS},
	"swagger-ui-bundle.js":            {"application/javascript; charset=utf-8", swaggerUIBundleJS},
	"swagger-ui-standalone-preset.js": {"application/javascript; charset=utf-8", swaggerUIStandalonePresetJS},
}

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

// ServeAsset serves one vendored swagger-ui-dist file.
//
// The :version segment is deliberately not compared against
// SwaggerUIDistVersion. The only caller that can supply an older one is a docs
// page we shipped ourselves and a browser still has open across an upgrade;
// answering it with the current bytes keeps that tab working, while a 404 would
// blank it. New loads always request the current version, so the immutable
// cache entry is never the stale one.
func (h *SwaggerHandler) ServeAsset(c echo.Context) error {
	asset, ok := swaggerUIAssets[c.Param("file")]
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	return c.Blob(http.StatusOK, asset.contentType, asset.body)
}
