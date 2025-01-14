package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
	webhookSecret  string
	webhookHandler *handler.WebhookHandler
	vpnHandler     *handler.VPNHandler
	peerHandler    *handler.PeerHandler
}

func (app *DonkeyVPNApplication) registerRoutes() {
	app.e.POST("/telegram/donkeyvpn/webhook", app.webhookHandler.Handle)
	app.e.GET("v1/api/vpn/nextid", app.vpnHandler.NextId)
	app.e.GET("v1/api/vpn/:vpnId", app.vpnHandler.Get)
	app.e.POST("v1/api/vpn/notify/:vpnId", app.vpnHandler.Notify)

	app.e.GET("v1/api/peer", app.peerHandler.List)
}

func (app *DonkeyVPNApplication) SecretBasedAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestToken := c.Request().Header.Get("x-api-key")
		if requestToken == "" {
			log.Warnf("x-api-key header is empty, testing with %f header", handler.TELEGRAM_WEBHOOK_SECRET_HEADER)
			requestToken = c.Request().Header.Get(handler.TELEGRAM_WEBHOOK_SECRET_HEADER)
		}
		if strings.Compare(requestToken, app.webhookSecret) != 0 {
			log.Warnf("received a missing or invalid webhook secret %v %v", requestToken, app.webhookSecret)
			c.Response().Header().Add("content-type", "text/plain")
			return c.JSON(http.StatusUnauthorized, "Invalid token, set the correct token in x-api-key header")
		}
		return next(c)
	}
}

func (app *DonkeyVPNApplication) Start() {
	app.e.Use(app.SecretBasedAuth)
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

	vpnService := service.NewVPNService(asg, instancesTable)
	peerService := service.NewWireguardPeerService(peersTable)

	cmdProcessor := processor.NewCommandProcessor()
	cmdProcessor.Register("/create", processor.NewCreateProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/list", processor.NewListProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/delete", processor.NewDeleteProcessor(client))
	cmdProcessor.RegisterFallback(processor.NewUnknowCommandProcessor(client))

	return &DonkeyVPNApplication{
		e:             e,
		webhookSecret: cfg.WebhookSecret,
		webhookHandler: &handler.WebhookHandler{
			WebhookSecret:    cfg.WebhookSecret,
			CommandProcessor: cmdProcessor,
		},
		vpnHandler: &handler.VPNHandler{
			WebhookSecret: cfg.WebhookSecret,
			VPNSvc:        vpnService,
		},
		peerHandler: &handler.PeerHandler{
			WebhookSecret: cfg.WebhookSecret,
			PeersTable:    peersTable,
		},
	}, nil
}
