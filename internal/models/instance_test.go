package models

import (
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

func TestChatIdValue(t *testing.T) {
	cases := map[string]telegram.ChatId{
		"12345":   12345,
		"-100123": -100123,
		"":        0,
	}
	for stored, expected := range cases {
		instance := VPNInstance{ChatId: stored}
		if got := instance.ChatIdValue(); got != expected {
			t.Errorf("ChatIdValue(%q) = %d, expected %d", stored, got, expected)
		}
	}
}
