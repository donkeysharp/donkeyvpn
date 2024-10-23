package telegram

import (
	"time"
)

type User struct {
	Id        uint32
	FirstName string
	LastName  string
	Username  string
}

type Chat struct {
	ChatId    uint32
	FirstName string
	LastName  string
	Username  string
	Type      string
}

type MessageEntities struct {
	Offset uint16
	Length uint32
	Type   string
}

type Message struct {
	MessageId uint32
	From      *User
	Chat      *Chat
	Date      time.Time
	Text      string
	Entities  []*MessageEntities
}

type Update struct {
	UpdateId uint32
	Message  *Message
}

func parseMessageEntities(entities []interface{}) []*MessageEntities {

	result := make([]*MessageEntities, 0)

	for i := 0; i < len(entities); i++ {
		entity := entities[i].(map[string]interface{})
		item := &MessageEntities{
			Offset: uint16(entity["offset"].(float64)),
			Length: uint32(entity["length"].(float64)),
			Type:   entity["type"].(string),
		}
		result = append(result, item)
	}

	return result
}

func NewUpdate(body map[string]interface{}) (*Update, error) {
	update := &Update{}
	update.UpdateId = uint32(body["update_id"].(float64))
	messageRaw := body["message"].(map[string]interface{})
	userRaw := messageRaw["from"].(map[string]interface{})

	user := &User{
		Id:       uint32(userRaw["id"].(float64)),
		Username: userRaw["username"].(string),
	}
	if val, ok := userRaw["fist_name"].(string); ok {
		user.LastName = val
	}
	if val, ok := userRaw["last_name"].(string); ok {
		user.LastName = val
	}

	chatRaw := messageRaw["chat"].(map[string]interface{})
	chat := &Chat{
		ChatId:   uint32(chatRaw["id"].(float64)),
		Username: chatRaw["username"].(string),
		Type:     chatRaw["type"].(string),
	}
	if val, ok := chatRaw["first_name"].(string); ok {
		chat.FirstName = val
	}
	if val, ok := chatRaw["last_name"].(string); ok {
		chat.LastName = val
	}

	entities := make([]*MessageEntities, 0)
	if messageRaw["entities"] != nil {
		entities = parseMessageEntities(messageRaw["entities"].([]interface{}))
	}

	update.Message = &Message{
		MessageId: uint32(messageRaw["message_id"].(float64)),
		From:      user,
		Date:      time.Unix(int64(messageRaw["date"].(float64)), 0),
		Chat:      chat,
		Text:      messageRaw["text"].(string),
		Entities:  entities,
	}

	return update, nil
}
