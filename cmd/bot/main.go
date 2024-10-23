package main

import (
	"os"

	"github.com/donkeysharp/donkeyvpn/internal/app"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func main() {
	var telegramBotAPIToken string = os.Getenv("TELEGRAM_BOT_API_TOKEN")
	var webhookSecret string = os.Getenv("WEBHOOK_SECRET")

	e := echo.New()
	app, err := app.NewApplication(app.DonkeyVPNConfig{
		TelegramBotAPIToken: telegramBotAPIToken,
		WebhookSecret:       webhookSecret,
	}, e)

	if err != nil {
		log.Error("error while creating a new DonkeyVPN application")
		os.Exit(1)
		return
	}
	app.Start()
}
