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

// DefaultBaseURL is the Telegram Bot API endpoint NewClient uses unless
// WithBaseURL overrides it.
const DefaultBaseURL = "https://api.telegram.org"

type ChatId int64

type replyMessage struct {
	ChatId    ChatId `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type Client struct {
	BotAPIToken string
	baseURL     string
	client      http.Client
}

// Option configures a Client created by NewClient.
type Option func(*Client)

// WithBaseURL sends API calls to baseURL instead of DefaultBaseURL, e.g. an
// httptest.Server in tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func (c *Client) methodURL(method string) string {
	return fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.BotAPIToken, method)
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
	url := c.methodURL("sendMessage")
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

func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		BotAPIToken: token,
		baseURL:     DefaultBaseURL,
		client:      http.Client{},
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}
