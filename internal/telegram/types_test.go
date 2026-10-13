package telegram

import (
	"testing"
)

func TestParseUpdate(t *testing.T) {
	cases := []struct {
		name        string
		body        string
		text        string
		chatId      ChatId
		unsupported bool
		invalid     bool
	}{
		{
			name: "text message",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"entities": [{"offset": 0, "length": 5, "type": "bot_command"}],
				"from": {"id": 42, "username": "alice"}, "chat": {"id": 42, "type": "private"}}}`,
			text:   "/list vpn",
			chatId: 42,
		},
		{
			name: "negative chat id",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42, "username": "alice"}, "chat": {"id": -100123, "type": "group"}}}`,
			text:   "/list vpn",
			chatId: -100123,
		},
		{
			name: "photo message defaults to help",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000,
				"photo": [{"file_id": "abc", "width": 90, "height": 90}],
				"from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`,
			text:   "/help",
			chatId: 42,
		},
		{
			name: "message without from",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"chat": {"id": -100123, "type": "channel"}}}`,
			text:   "/list vpn",
			chatId: -100123,
		},
		{
			name: "edited_message",
			body: `{"update_id": 1, "edited_message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}, "chat": {"id": 42, "type": "private"}}}`,
			unsupported: true,
		},
		{
			name:        "callback_query",
			body:        `{"update_id": 1, "callback_query": {"id": "abc", "from": {"id": 42}, "data": "delete:vpn001"}}`,
			unsupported: true,
		},
		{
			name:        "my_chat_member",
			body:        `{"update_id": 1, "my_chat_member": {"chat": {"id": 42, "type": "private"}}}`,
			unsupported: true,
		},
		{
			name:    "invalid json",
			body:    `Empty`,
			invalid: true,
		},
		{
			name:    "json that is not an object",
			body:    `[1, 2]`,
			invalid: true,
		},
		{
			name:    "null body",
			body:    `null`,
			invalid: true,
		},
		{
			name: "missing chat.id",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}, "chat": {"type": "private"}}}`,
			invalid: true,
		},
		{
			name: "missing chat",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000, "text": "/list vpn",
				"from": {"id": 42}}}`,
			invalid: true,
		},
		{
			name:    "message with the wrong type",
			body:    `{"update_id": 1, "message": "hello"}`,
			invalid: true,
		},
		{
			name: "chat.id with the wrong type",
			body: `{"update_id": 1, "message": {"message_id": 2, "date": 1700000000,
				"chat": {"id": "42", "type": "private"}}}`,
			invalid: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			update, err := ParseUpdate([]byte(tc.body))
			if tc.unsupported {
				if err != ErrUnsupportedUpdate {
					t.Fatalf("expected ErrUnsupportedUpdate, got %v", err)
				}
				return
			}
			if tc.invalid {
				if err == nil || err == ErrUnsupportedUpdate {
					t.Fatalf("expected a parse error, got %v", err)
				}
				if update != nil {
					t.Errorf("expected no update, got %+v", update)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if update.Message.Text != tc.text {
				t.Errorf("expected text %q, got %q", tc.text, update.Message.Text)
			}
			if update.Message.Chat.ChatId != tc.chatId {
				t.Errorf("expected chat id %d, got %d", tc.chatId, update.Message.Chat.ChatId)
			}
			if update.Message.From == nil {
				t.Error("expected From to be set")
			}
		})
	}
}

func TestParseUpdateFillsAllFields(t *testing.T) {
	body := `{"update_id": 7, "message": {"message_id": 8, "date": 1700000000, "text": "/create vpn",
		"entities": [{"offset": 0, "length": 7, "type": "bot_command"}],
		"from": {"id": 42, "first_name": "Alice", "last_name": "Doe", "username": "alice"},
		"chat": {"id": 43, "first_name": "Bob", "last_name": "Roe", "username": "bob", "type": "private"}}}`

	update, err := ParseUpdate([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if update.UpdateId != 7 || update.Message.MessageId != 8 {
		t.Errorf("unexpected ids: update=%d message=%d", update.UpdateId, update.Message.MessageId)
	}
	if update.Message.Date.Unix() != 1700000000 {
		t.Errorf("unexpected date %v", update.Message.Date)
	}
	from := update.Message.From
	if from.Id != 42 || from.FirstName != "Alice" || from.LastName != "Doe" || from.Username != "alice" {
		t.Errorf("unexpected user %+v", from)
	}
	chat := update.Message.Chat
	if chat.ChatId != 43 || chat.FirstName != "Bob" || chat.LastName != "Roe" || chat.Username != "bob" || chat.Type != "private" {
		t.Errorf("unexpected chat %+v", chat)
	}
	entities := update.Message.Entities
	if len(entities) != 1 || entities[0].Offset != 0 || entities[0].Length != 7 || entities[0].Type != "bot_command" {
		t.Errorf("unexpected entities %+v", entities)
	}
}
