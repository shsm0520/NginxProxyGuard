package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"nginx-proxy-guard/internal/model"
)

func filterCtx(t *testing.T, rawQuery string) echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/proxy-hosts?"+rawQuery, nil)
	return e.NewContext(req, httptest.NewRecorder())
}

func TestParseProxyHostListFilter(t *testing.T) {
	f, err := parseProxyHostListFilter(filterCtx(t, "tag=Media&tag=family&tag=media&domain=example.com&upstream=192.0.2.9&enabled=true"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Tags) != 2 || f.Tags[0] != "media" || f.Tags[1] != "family" {
		t.Fatalf("tags normalised+deduped wrong: %v", f.Tags)
	}
	if f.Domain != "example.com" || f.Upstream != "192.0.2.9" || f.Enabled == nil || !*f.Enabled {
		t.Fatalf("filter = %+v", f)
	}

	empty, err := parseProxyHostListFilter(filterCtx(t, ""))
	if err != nil || empty.Tags != nil || empty.Enabled != nil || empty.Domain != "" {
		t.Fatalf("empty query must yield a zero filter, got %+v (%v)", empty, err)
	}

	for _, bad := range []string{"tag=-bad", "enabled=maybe", "domain=two%20words"} {
		if _, err := parseProxyHostListFilter(filterCtx(t, bad)); !errors.Is(err, model.ErrInvalidInput) {
			t.Errorf("%q: want ErrInvalidInput, got %v", bad, err)
		}
	}
}
