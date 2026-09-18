package handler

import (
	"strings"
	"testing"
)

// The docs page and the routes that feed it agree on a version string that
// appears in two places — SwaggerUIDistVersion and the three URLs inside
// swagger_ui.html. Bumping one and not the other produces a page that loads a
// 404 for its own script, which looks exactly like the CDN outage this vendoring
// was meant to end. Catch it at build time instead.
func TestSwaggerUIHTMLReferencesTheVendoredAssets(t *testing.T) {
	for name, asset := range swaggerUIAssets {
		if len(asset.body) == 0 {
			t.Errorf("vendored asset %s is empty — the go:embed did not pick up a file", name)
		}
		want := "/api/docs/assets/" + SwaggerUIDistVersion + "/" + name
		if !strings.Contains(swaggerUIHTML, want) {
			t.Errorf("swagger_ui.html does not reference %q; bump the URLs in swagger_ui.html "+
				"to match SwaggerUIDistVersion (%s)", want, SwaggerUIDistVersion)
		}
	}

	// The whole point of the change: no third-party origin, and no online
	// validator badge (the API's CSP was tightened to match both).
	for _, external := range []string{`src="http`, `href="http`} {
		if strings.Contains(swaggerUIHTML, external) {
			t.Errorf("swagger_ui.html has an %s… subresource — /api/docs must render with no egress, "+
				"and the API's CSP no longer allows an off-box origin", external)
		}
	}
	if !strings.Contains(swaggerUIHTML, "validatorUrl: null") {
		t.Error("swagger_ui.html no longer disables validatorUrl — swagger-ui defaults it to " +
			"https://validator.swagger.io/validator, which the API's img-src no longer allows")
	}
}

// The route the assets are served on must stay in the shape PublicRoutes and
// ServeAsset agree on: two parameters, the second named "file".
func TestSwaggerAssetsRouteShape(t *testing.T) {
	if want := "/api/docs/assets/:version/:file"; SwaggerAssetsRoute != want {
		t.Fatalf("SwaggerAssetsRoute = %q, want %q — middleware/permissions.go carries this "+
			"pattern verbatim as a PublicRoutes key", SwaggerAssetsRoute, want)
	}
}
