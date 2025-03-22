package main

import (
	"os"

	"github.com/donkeysharp/donkeyvpn/internal/app"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	var telegramBotAPIToken string = os.Getenv("TELEGRAM_BOT_API_TOKEN")
	var webhookSecret string = os.Getenv("WEBHOOK_SECRET")
	var autoscalingGroupName string = os.Getenv("AUTOSCALING_GROUP_NAME")
	var peersTableName string = os.Getenv("DYNAMODB_PEERS_TABLE_NAME")
	var instancesTableName string = os.Getenv("DYNAMODB_INSTANCES_TABLE_NAME")
	var runAsLambdaStr string = os.Getenv("RUN_AS_LAMBDA")
	var runAsLambda bool = false
	if runAsLambdaStr == "true" {
		runAsLambda = true
	}

	e := echo.New()
	app, err := app.NewApplication(config.DonkeyVPNConfig{
		TelegramBotAPIToken:  telegramBotAPIToken,
		WebhookSecret:        webhookSecret,
		AutoscalingGroupName: autoscalingGroupName,
		PeersTableName:       peersTableName,
		InstancesTableName:   instancesTableName,
		RunAsLambda:          runAsLambda,
	}, e)

	if err != nil {
		log.Error("error while creating a new DonkeyVPN application")
		os.Exit(1)
		return
	}
	app.Start()
}
