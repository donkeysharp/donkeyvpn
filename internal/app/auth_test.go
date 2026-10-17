package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	configuredSecret = "configured-secret-value"
	suppliedToken    = "supplied-token-value"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Output()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prev) })
	return &buf
}

func TestSecretBasedAuthDoesNotLogSecrets(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		header string
	}{
		{name: "wrong x-api-key", path: "/v1/api/peer", header: "x-api-key"},
		{name: "wrong telegram secret", path: "/v1/api/telegram/donkeyvpn/webhook", header: handler.TELEGRAM_WEBHOOK_SECRET_HEADER},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logs := captureLog(t)
			app := &DonkeyVPNApplication{e: echo.New(), webhookSecret: configuredSecret}
			app.registerRoutes()

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.path == "/v1/api/telegram/donkeyvpn/webhook" {
				req.Method = http.MethodPost
			}
			req.Header.Set(tc.header, suppliedToken)
			rec := httptest.NewRecorder()
			app.e.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected status 401, got %d", rec.Code)
			}
			out := logs.String()
			if !strings.Contains(out, "invalid or missing secret") || !strings.Contains(out, tc.path) {
				t.Errorf("expected the failure to be logged with the path, got %q", out)
			}
			if strings.Contains(out, configuredSecret) || strings.Contains(out, suppliedToken) {
				t.Errorf("log output contains a secret: %q", out)
			}
		})
	}
}

func TestSecretBasedAuthAcceptsValidSecret(t *testing.T) {
	for _, header := range []string{"x-api-key", handler.TELEGRAM_WEBHOOK_SECRET_HEADER} {
		t.Run(header, func(t *testing.T) {
			app := &DonkeyVPNApplication{webhookSecret: configuredSecret}
			e := echo.New()
			e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, app.SecretBasedAuth)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(strings.ToLower(header), configuredSecret)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}
		})
	}
}
