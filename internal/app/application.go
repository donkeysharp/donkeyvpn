package app

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type DonkeyVPNConfig struct {
	TelegramBotAPIToken string
	WebhookSecret       string
}

type DonkeyVPNApplication struct {
	e              *echo.Echo
	webhookHandler *handler.WebhookHandler
}

func (app *DonkeyVPNApplication) registerRoutes() {
	app.e.POST("/telegram/donkeyvpn/webhook", app.webhookHandler.Handle)
}

func (app *DonkeyVPNApplication) Start() {
	app.registerRoutes()
	app.e.Logger.Fatal(app.e.Start(":8080"))
}

func NewApplication(cfg DonkeyVPNConfig, e *echo.Echo) (*DonkeyVPNApplication, error) {
	if cfg.TelegramBotAPIToken == "" {
		msg := "missing telegram bot API token"
		log.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	if cfg.WebhookSecret == "" {
		msg := "missing webhook secret config"
		log.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	client := telegram.NewClient(cfg.TelegramBotAPIToken)

	svc := service.NewCommandService()
	svc.Register("/create", processor.NewCreateProcessor(client))
	svc.Register("/list", processor.NewListProcessor(client))
	svc.Register("/delete", processor.NewDeleteProcessor(client))
	svc.RegisterFallback(processor.NewUnknowCommandProcessor(client))

	return &DonkeyVPNApplication{
		e: e,
		webhookHandler: &handler.WebhookHandler{
			WebhookSecret:  cfg.WebhookSecret,
			CommandService: svc,
		},
	}, nil
}
