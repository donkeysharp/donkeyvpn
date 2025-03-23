package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type DonkeyVPNApplication struct {
	e              *echo.Echo
	runAsLambda    bool
	webhookSecret  string
	webhookHandler *handler.WebhookHandler
	vpnHandler     *handler.VPNHandler
	peerHandler    *handler.PeerHandler
}

var echoLambda *echoadapter.EchoLambdaV2

func (app *DonkeyVPNApplication) registerRoutes() {
	api := app.e.Group("/v1/api")

	// Authentication
	api.Use(app.SecretBasedAuth)
	// Telegram bot will send all messages to this endpoint
	api.POST("/telegram/donkeyvpn/webhook", app.webhookHandler.Handle)
	// Used by user-data script to retrieve the vpn instance that is being created
	api.GET("/vpn/pending", app.vpnHandler.GetPendingId)
	// Used by user-data script to notify when everything has been created or not
	api.POST("/vpn/notify/:vpnId", app.vpnHandler.Notify)
	// Used for validation
	api.GET("/peer", app.peerHandler.List)

	// public endpoints
	app.e.GET("/donkeyvpn/ping", app.vpnHandler.Ping)
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

func (app *DonkeyVPNApplication) HandlerV2(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}

func (app *DonkeyVPNApplication) Start() {
	app.registerRoutes()

	if !app.runAsLambda {
		log.Info("Running as like executable application")
		app.e.Logger.Fatal(app.e.Start(":8080"))
	} else {
		// Starting lambda handler with compatibility to Echo
		log.Info("Running as a Lambda handler")
		echoLambda = echoadapter.NewV2(app.e)
		lambda.Start(app.HandlerV2)
	}

}

func NewApplication(cfg config.DonkeyVPNConfig, e *echo.Echo) (*DonkeyVPNApplication, error) {
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
	peerService := service.NewWireguardPeerService(peersTable, cfg.WireguardCidrRange)

	cmdProcessor := processor.NewCommandProcessor()
	cmdProcessor.Register("/create", processor.NewCreateProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/list", processor.NewListProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/delete", processor.NewDeleteProcessor(client, vpnService, peerService))
	cmdProcessor.RegisterFallback(processor.NewUnknowCommandProcessor(client))

	return &DonkeyVPNApplication{
		e:             e,
		runAsLambda:   cfg.RunAsLambda,
		webhookSecret: cfg.WebhookSecret,
		webhookHandler: &handler.WebhookHandler{
			WebhookSecret:    cfg.WebhookSecret,
			CommandProcessor: cmdProcessor,
		},
		vpnHandler: &handler.VPNHandler{
			WebhookSecret:  cfg.WebhookSecret,
			VPNSvc:         vpnService,
			TelegramClient: client,
		},
		peerHandler: &handler.PeerHandler{
			WebhookSecret: cfg.WebhookSecret,
			PeersTable:    peersTable,
		},
	}, nil
}
