package handler

import (
	"crypto/subtle"
	"io"
	"net/http"

	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

const (
	TELEGRAM_WEBHOOK_SECRET_HEADER = "X-Telegram-Bot-Api-Secret-Token"
	GenericErrorMessage            = "Error processing request."
)

type WebhookHandler struct {
	WebhookSecret    string
	CommandProcessor *processor.CommandProcessor
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
	if subtle.ConstantTimeCompare([]byte(requestToken), []byte(h.WebhookSecret)) != 1 {
		log.Errorf("invalid or missing secret in %s header for %s", TELEGRAM_WEBHOOK_SECRET_HEADER, c.Request().URL.Path)
		c.Response().Header().Add("content-type", "text/plain")
		return c.String(http.StatusAccepted, GenericErrorMessage)
	}

	log.Infof("Body content: %s", content)

	update, err := telegram.ParseUpdate(contentRaw)
	if err == telegram.ErrUnsupportedUpdate {
		log.Info("ignoring update without a message")
		return c.String(http.StatusOK, "Update ignored.")
	}
	if err != nil {
		log.Errorf("error parsing body to create Update struct: %v", err)
		return c.String(http.StatusOK, GenericErrorMessage)
	}

	log.Info("sending answer to message")
	h.CommandProcessor.Process(update)

	return c.String(http.StatusOK, content)
}
