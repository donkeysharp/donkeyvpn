package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/donkeysharp/donkeyvpn/internal/testutil"
	"github.com/labstack/echo/v5"
)

func notify(t *testing.T, h *VPNHandler, vpnId, body string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.POST("/v1/api/vpn/notify/:vpnId", h.Notify)
	req := httptest.NewRequest(http.MethodPost, "/v1/api/vpn/notify/"+vpnId, strings.NewReader(body))
	req.Header.Set("content-type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

const notifyBody = `{"hostname":"vpn.example.com","port":"51820","status":"success","instanceId":"i-123"}`

func TestNotifyUnknownVPNReturns404(t *testing.T) {
	table := testutil.NewFakeTable()
	table.CreateRecord(models.NewVPNInstance("vpn001", "pending", "pending", "pending", "pending", "42"))
	sender := &testutil.RecordingSender{}
	h := &VPNHandler{VPNSvc: service.NewVPNService(nil, table, nil), TelegramClient: sender}

	rec := notify(t, h, "vpn009", notifyBody)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got["message"] != "VPN instance not found" {
		t.Errorf("unexpected body %q", rec.Body.String())
	}
	if table.Len() != 1 {
		t.Errorf("expected no record to be created, got %d records", table.Len())
	}
	if msgs := sender.Messages(); len(msgs) != 0 {
		t.Errorf("expected no Telegram message, got %v", msgs)
	}
}

func TestNotifyUpdatesRecordAndMessagesChat(t *testing.T) {
	table := testutil.NewFakeTable()
	table.CreateRecord(models.NewVPNInstance("vpn001", "pending", "pending", "pending", "pending", "-100123"))
	sender := &testutil.RecordingSender{}
	svc := service.NewVPNService(nil, table, nil)
	h := &VPNHandler{VPNSvc: svc, TelegramClient: sender}

	rec := notify(t, h, "vpn001", notifyBody)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", rec.Code, rec.Body.String())
	}
	got, err := svc.Get("vpn001")
	if err != nil {
		t.Fatal(err)
	}
	if want := *models.NewVPNInstance("vpn001", "vpn.example.com", "51820", "success", "i-123", "-100123"); *got != want {
		t.Errorf("expected record %v, got %v", want, got)
	}
	msgs := sender.Messages()
	want := testutil.SentMessage{ChatId: telegram.ChatId(-100123), Text: "✅ VPN Instance with id `vpn001` provisioned *successfully*"}
	if len(msgs) != 1 || msgs[0] != want {
		t.Errorf("expected %v, got %v", want, msgs)
	}
}
