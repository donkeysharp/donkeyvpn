package app

import (
	"context"
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type DonkeyVPNConfig struct {
	TelegramBotAPIToken  string
	WebhookSecret        string
	AutoscalingGroupName string
	PeersTableName       string
	InstancesTableName   string
}

type DonkeyVPNApplication struct {
	e              *echo.Echo
	webhookHandler *handler.WebhookHandler
	vpnHandler     *handler.VPNHandler
}

func (app *DonkeyVPNApplication) registerRoutes() {
	app.e.POST("/telegram/donkeyvpn/webhook", app.webhookHandler.Handle)
	app.e.POST("v1/api/vpn", app.vpnHandler.Handle)
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

	ctx := context.Background()
	asg, err := aws.NewAutoscalingGroup(ctx, cfg.AutoscalingGroupName)
	if err != nil {
		return nil, err
	}

	peersTable, err := aws.NewDynamoDB(ctx, cfg.PeersTableName)
	if err != nil {
		return nil, err
	}

	instancesTable, err := aws.NewDynamoDB(ctx, cfg.InstancesTableName)
	if err != nil {
		return nil, err
	}

	svc := service.NewCommandService()
	svc.Register("/create", processor.NewCreateProcessor(client, asg, peersTable))
	svc.Register("/list", processor.NewListProcessor(client, peersTable, instancesTable))
	svc.Register("/delete", processor.NewDeleteProcessor(client))
	svc.RegisterFallback(processor.NewUnknowCommandProcessor(client))

	return &DonkeyVPNApplication{
		e: e,
		webhookHandler: &handler.WebhookHandler{
			WebhookSecret:  cfg.WebhookSecret,
			CommandService: svc,
		},
		vpnHandler: &handler.VPNHandler{
			WebhookSecret:  cfg.WebhookSecret,
			InstancesTable: instancesTable,
		},
	}, nil
}
