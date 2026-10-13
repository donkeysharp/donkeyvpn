package config

import (
	"reflect"
	"testing"
)

func TestParseAllowedChatIds(t *testing.T) {
	got := ParseAllowedChatIds("12345, -100123 ,abc")
	expected := []int64{12345, -100123}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestParseAllowedChatIdsEmpty(t *testing.T) {
	for _, raw := range []string{"", "  ", ","} {
		if got := ParseAllowedChatIds(raw); len(got) != 0 {
			t.Errorf("ParseAllowedChatIds(%q) = %v, expected empty", raw, got)
		}
	}
}

func TestLoadConfigFromEnvVarsAllowedChatIdsSSMParam(t *testing.T) {
	t.Setenv("SSM_ALLOWED_CHAT_IDS", "/donkeyvpn/allowedchatids")
	cfg := LoadConfigFromEnvVars()
	if cfg.AllowedChatIdsSSMParam != "/donkeyvpn/allowedchatids" {
		t.Errorf("unexpected SSM param %q", cfg.AllowedChatIdsSSMParam)
	}
}
