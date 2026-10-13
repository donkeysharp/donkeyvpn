package telegram

import (
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	Id        ChatId `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Chat struct {
	ChatId    ChatId `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Type      string `json:"type"`
}

type MessageEntities struct {
	Offset uint16 `json:"offset"`
	Length uint32 `json:"length"`
	Type   string `json:"type"`
}

type Message struct {
	MessageId uint32             `json:"message_id"`
	From      *User              `json:"from"`
	Chat      *Chat              `json:"chat"`
	Date      time.Time          `json:"date"`
	Text      string             `json:"text"`
	Entities  []*MessageEntities `json:"entities"`
}

type Update struct {
	UpdateId uint32   `json:"update_id"`
	Message  *Message `json:"message"`
}

var ErrUnsupportedUpdate = fmt.Errorf("update does not contain a message")

// for custom fields such as date and empty text
func (m *Message) UnmarshalJSON(data []byte) error {
	type plain Message
	aux := struct {
		*plain
		Date int64   `json:"date"`
		Text *string `json:"text"`
	}{plain: (*plain)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Date = time.Unix(aux.Date, 0)
	m.Text = "/help" // defaults to /help
	if aux.Text != nil {
		m.Text = *aux.Text
	}
	// from is empty for messages sent to channels
	if m.From == nil {
		m.From = &User{}
	}
	if m.Entities == nil {
		m.Entities = make([]*MessageEntities, 0)
	}
	return nil
}

func ParseUpdate(body []byte) (*Update, error) {
	update := &Update{}
	if err := json.Unmarshal(body, update); err != nil {
		return nil, fmt.Errorf("invalid update: %w", err)
	}
	if update.UpdateId == 0 {
		return nil, fmt.Errorf("invalid update: missing update_id")
	}
	if update.Message == nil {
		return nil, ErrUnsupportedUpdate
	}
	if update.Message.Chat == nil || update.Message.Chat.ChatId == 0 {
		return nil, fmt.Errorf("invalid update: missing message.chat.id")
	}
	return update, nil
}
