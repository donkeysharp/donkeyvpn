package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/gommon/log"
)

const (
	BASE_URL = "https://api.telegram.org/bot%s"
)

func getBaseUrl(token string) string {
	return fmt.Sprintf(BASE_URL, token)
}

type replyMessage struct {
	ChatId uint32 `json:"chat_id"`
	Text   string `json:"text"`
}

type Client struct {
	BotAPIToken string
	client      http.Client
}

func (c *Client) SendMessage(message string, chat *Chat) error {
	reply := replyMessage{ChatId: chat.ChatId, Text: message}
	buf, err := json.Marshal(reply)
	if err != nil {
		log.Error("Could not marshal replyMessage struct")
		return err
	}
	url := getBaseUrl(c.BotAPIToken + "/sendMessage")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		log.Error("Error while creating sendMessage request")
		return err
	}
	req.Header.Add("Content-type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Error("Error while calling to sendMessage API")
		return err
	}
	log.Infof("sendMessage status code: %d", resp.StatusCode)
	return nil
}

func (c *Client) SendErrorMessage(chat *Chat) error {
	log.Error("Sending default error message.")
	message := "Sorry, something went wrong while processing your request. Try again please."
	return c.SendMessage(message, chat)
}

func NewClient(token string) *Client {
	c := &Client{
		BotAPIToken: token,
		client:      http.Client{},
	}

	return c
}
