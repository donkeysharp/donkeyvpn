package telegram

import (
	"encoding/json"
	"testing"
)

func TestNewUpdateParsesNegativeChatId(t *testing.T) {
	raw := `{
		"update_id": 1,
		"message": {
			"message_id": 2,
			"date": 1700000000,
			"text": "/list vpn",
			"from": {"id": 42, "username": "alice"},
			"chat": {"id": -100123, "type": "group"}
		}
	}`
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}

	update, err := NewUpdate(body)
	if err != nil {
		t.Fatal(err)
	}
	if update.Message.Chat.ChatId != ChatId(-100123) {
		t.Errorf("expected chat id -100123, got %d", update.Message.Chat.ChatId)
	}
}
