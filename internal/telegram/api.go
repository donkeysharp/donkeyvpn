package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/gommon/log"
)

const (
	BASE_URL = "https://api.telegram.org/bot%s"
)

func getBaseUrl(token string) string {
	return fmt.Sprintf(BASE_URL, token)
}

type ChatId uint64

type replyMessage struct {
	ChatId    ChatId `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type Client struct {
	BotAPIToken string
	client      http.Client
}

func escapeMessage(message string) string {
	// _*`~ won't be escaped as they are used
	escapeChars := "[]()>#+-=|{}.!"

	newMessage := message
	for _, char := range escapeChars {
		newMessage = strings.ReplaceAll(newMessage, string(char), "\\"+string(char))
	}
	return newMessage
}

func (c *Client) SendMessage(message string, chat *Chat) error {
	escapedMessage := escapeMessage(message)
	reply := replyMessage{ChatId: chat.ChatId, Text: escapedMessage, ParseMode: "MarkdownV2"}
	buf, err := json.Marshal(reply)
	if err != nil {
		log.Error("Could not marshal replyMessage struct")
		return err
	}
	url := getBaseUrl(c.BotAPIToken + "/sendMessage")
	log.Infof("Sending message to telegram. ChatId: %v Message: %v", chat.ChatId, escapedMessage)
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
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Error while reading sendMessage response body %v", err.Error())
		return nil
	}
	log.Infof("sendMessage response %v", string(respBody))
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
