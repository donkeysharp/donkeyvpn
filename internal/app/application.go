package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	"github.com/donkeysharp/donkeyvpn/internal/handler"
	"github.com/donkeysharp/donkeyvpn/internal/lambdahttp"
	"github.com/donkeysharp/donkeyvpn/internal/processor"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
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

const apiPrefix = "/v1/api"

func (app *DonkeyVPNApplication) registerRoutes() {
	// A handler panic becomes a 500 instead of failing the Lambda invocation
	app.e.Use(middleware.Recover())

	// Authentication for /v1/api. It is registered at the root rather than on
	// the group: group middleware makes Echo add a catch-all route, which turns
	// a wrong method on a known path into 404 instead of 405.
	app.e.Use(app.SecretBasedAuth)

	api := app.e.Group(apiPrefix)
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
	return func(c *echo.Context) error {
		if path := c.Request().URL.Path; path != apiPrefix && !strings.HasPrefix(path, apiPrefix+"/") {
			return next(c)
		}
		header := "x-api-key"
		requestToken := c.Request().Header.Get(header)
		if requestToken == "" {
			header = handler.TELEGRAM_WEBHOOK_SECRET_HEADER
			requestToken = c.Request().Header.Get(header)
		}
		if subtle.ConstantTimeCompare([]byte(requestToken), []byte(app.webhookSecret)) != 1 {
			log.Warnf("invalid or missing secret in %s header for %s", header, c.Request().URL.Path)
			c.Response().Header().Add("content-type", "text/plain")
			return c.JSON(http.StatusUnauthorized, "Invalid token, set the correct token in x-api-key header")
		}
		return next(c)
	}
}

func (app *DonkeyVPNApplication) Start() {
	app.registerRoutes()

	if !app.runAsLambda {
		log.Info("Running as like executable application")
		if err := app.e.Start(":8080"); err != nil {
			log.Fatalf("http server stopped: %v", err)
		}
	} else {
		log.Info("Running as a Lambda handler")
		lambda.Start(lambdahttp.Handler(app.e))
	}

}

func NewApplication(cfg config.DonkeyVPNConfig, e *echo.Echo) (*DonkeyVPNApplication, error) {
	if cfg.TelegramBotAPIToken == "" {
		msg := "missing telegram bot API token"
		log.Error(msg)
		return nil, errors.New(msg)
	}

	if cfg.WebhookSecret == "" {
		msg := "missing webhook secret config"
		log.Error(msg)
		return nil, errors.New(msg)
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

	ssmClient, err := aws.NewSSM(ctx)
	if err != nil {
		return nil, err
	}

	vpnService := service.NewVPNService(asg, instancesTable, nil)
	peerService := service.NewWireguardPeerService(peersTable, cfg.WireguardCidrRange)

	authorizer := processor.NewAuthorizer(func() ([]int64, error) {
		if cfg.AllowedChatIdsSSMParam == "" {
			return nil, fmt.Errorf("SSM_ALLOWED_CHAT_IDS is not set")
		}
		value, err := ssmClient.GetParameter(cfg.AllowedChatIdsSSMParam, true)
		if err != nil {
			return nil, err
		}
		return config.ParseAllowedChatIds(value), nil
	})
	cmdProcessor := processor.NewCommandProcessor(authorizer, processor.NewUnauthorizedProcessor(client))
	cmdProcessor.Register("/create", processor.NewCreateProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/list", processor.NewListProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/delete", processor.NewDeleteProcessor(client, vpnService, peerService))
	cmdProcessor.Register("/settings", processor.NewSettingsProcessor(client, ssmClient, &cfg))
	cmdProcessor.Register("/docs", processor.NewDocsProcessor(client))
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
