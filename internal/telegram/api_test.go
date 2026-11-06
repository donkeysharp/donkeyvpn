package telegram

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientDefaultsToTelegramAPI(t *testing.T) {
	c := NewClient("123:abc")
	if got, want := c.methodURL("sendMessage"), "https://api.telegram.org/bot123:abc/sendMessage"; got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestSendMessage(t *testing.T) {
	var gotMethod, gotPath, gotContentType string
	var gotBody replyMessage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotContentType = r.Method, r.URL.Path, r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("request body is not JSON: %q", raw)
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	c := NewClient("123:abc", WithBaseURL(server.URL+"/"))
	if err := c.SendMessage("✅ VPN vpn001 is ready.", &Chat{ChatId: -100123}); err != nil {
		t.Fatal(err)
	}

	if gotMethod != http.MethodPost || gotPath != "/bot123:abc/sendMessage" {
		t.Errorf("expected POST /bot123:abc/sendMessage, got %s %s", gotMethod, gotPath)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected application/json, got %q", gotContentType)
	}
	want := replyMessage{ChatId: -100123, Text: `✅ VPN vpn001 is ready\.`, ParseMode: "MarkdownV2"}
	if gotBody != want {
		t.Errorf("expected body %+v, got %+v", want, gotBody)
	}
}

func TestEscapeMessage(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{name: "every escaped char", in: "[]()>#+-=|{}.!", want: `\[\]\(\)\>\#\+\-\=\|\{\}\.\!`},
		{name: "formatting chars untouched", in: "*bold* _italic_ `code` ~strike~", want: "*bold* _italic_ `code` ~strike~"},
		{name: "pre-escaped underscore kept", in: `<peer\_ip>`, want: `<peer\_ip\>`},
		{name: "ip and hostname", in: "*Hostname*: vpn-1.example.com (10.0.0.1)", want: `*Hostname*: vpn\-1\.example\.com \(10\.0\.0\.1\)`},
		{name: "emoji and plain text", in: "✅ Peer deleted", want: "✅ Peer deleted"},
		{name: "empty", in: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeMessage(tc.in); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
