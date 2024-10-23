package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	TELEGRAM_WEBHOOK_SECRET_HEADER = "X-Telegram-Bot-Api-Secret-Token"
	GenericErrorMessage            = "Error processing request."
)

type WebhookHandler struct {
	WebhookSecret  string
	CommandService *service.CommandService
}

func NewWebhookHandler(webhookSecret string, commandService *service.CommandService) *WebhookHandler {
	return &WebhookHandler{
		WebhookSecret:  webhookSecret,
		CommandService: commandService,
	}
}

func (h *WebhookHandler) Handle(c echo.Context) error {
	contentRaw, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Error("could not load body reader")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}

	if len(contentRaw) == 0 {
		contentRaw = []byte("Empty")
	}
	c.Response().Header().Add("Content-type", "application/json")
	content := string(contentRaw)
	requestToken := c.Request().Header.Get(TELEGRAM_WEBHOOK_SECRET_HEADER)
	if strings.Compare(requestToken, h.WebhookSecret) != 0 {
		log.Error("received a missing or invalid webhook secret")
		c.Response().Header().Add("content-type", "text/plain")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}

	log.Infof("Body content: %s", content)

	var tmp interface{}
	err = json.Unmarshal(contentRaw, &tmp)
	if err != nil {
		log.Error("error unmarshalling request body")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}
	updateRaw := tmp.(map[string]interface{})

	update, err := telegram.NewUpdate(updateRaw)
	if err != nil {
		log.Error("error parsing body to create Update struct")
		return c.String(http.StatusInternalServerError, GenericErrorMessage)
	}

	fmt.Printf("%+v\n", update)
	log.Info("sending answer to message")
	h.CommandService.Process(update)

	return c.String(http.StatusOK, content)
}
