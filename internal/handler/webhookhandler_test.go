package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const testWebhookSecret = "test-secret"

type recordingProcessor struct {
	calls int
}

func (p *recordingProcessor) Process(args []string, update *telegram.Update) error {
	p.calls++
	return nil
}

func TestWebhookHandleDoesNotLogSecrets(t *testing.T) {
	const suppliedToken = "supplied-token-value"
	var logs bytes.Buffer
	prev := log.Output()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(prev) })

	list := &recordingProcessor{}
	cmd := processor.NewCommandProcessor(processor.NewAuthorizer(func() ([]int64, error) { return []int64{42}, nil }), list)
	cmd.RegisterFallback(list)
	h := &WebhookHandler{WebhookSecret: testWebhookSecret, CommandProcessor: cmd}

	body := `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
		"from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/api/telegram/donkeyvpn/webhook", strings.NewReader(body))
	req.Header.Set(TELEGRAM_WEBHOOK_SECRET_HEADER, suppliedToken)
	rec := httptest.NewRecorder()

	if err := h.Handle(echo.New().NewContext(req, rec)); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", rec.Code)
	}
	if list.calls != 0 {
		t.Errorf("expected no command to run, got %d calls", list.calls)
	}
	out := logs.String()
	if !strings.Contains(out, "invalid or missing secret") {
		t.Errorf("expected the failure to be logged, got %q", out)
	}
	if strings.Contains(out, testWebhookSecret) || strings.Contains(out, suppliedToken) {
		t.Errorf("log output contains a secret: %q", out)
	}
}

func TestWebhookHandle(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		list     int
		fallback int
	}{
		{
			name: "command message",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`,
			list: 1,
		},
		{
			name: "message without text",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000,
				"photo": [{"file_id": "abc"}], "from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`,
			fallback: 1,
		},
		{
			name: "edited message",
			body: `{"update_id": 1, "edited_message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`,
		},
		{
			name: "unsupported update type",
			body: `{"update_id": 1, "my_chat_member": {"chat": {"id": 42, "type": "private"}}}`,
		},
		{
			name: "callback query",
			body: `{"update_id": 1, "callback_query": {"id": "abc", "from": {"id": 42}, "data": "x"}}`,
		},
		{
			name: "invalid json",
			body: `{"update_id": 1, "message": `,
		},
		{
			name: "empty body",
			body: ``,
		},
		{
			name: "missing chat id",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}, "chat": {"type": "private"}}}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list := &recordingProcessor{}
			fallback := &recordingProcessor{}
			unauthorized := &recordingProcessor{}
			loader := func() ([]int64, error) { return []int64{42}, nil }
			cmd := processor.NewCommandProcessor(processor.NewAuthorizer(loader), unauthorized)
			cmd.Register("/list", list)
			cmd.RegisterFallback(fallback)
			h := &WebhookHandler{WebhookSecret: testWebhookSecret, CommandProcessor: cmd}

			req := httptest.NewRequest(http.MethodPost, "/v1/api/telegram/donkeyvpn/webhook", strings.NewReader(tc.body))
			req.Header.Set(TELEGRAM_WEBHOOK_SECRET_HEADER, testWebhookSecret)
			rec := httptest.NewRecorder()

			if err := h.Handle(echo.New().NewContext(req, rec)); err != nil {
				t.Fatal(err)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", rec.Code)
			}
			if list.calls != tc.list || fallback.calls != tc.fallback || unauthorized.calls != 0 {
				t.Errorf("expected list=%d fallback=%d unauthorized=0, got list=%d fallback=%d unauthorized=%d",
					tc.list, tc.fallback, list.calls, fallback.calls, unauthorized.calls)
			}
		})
	}
}
