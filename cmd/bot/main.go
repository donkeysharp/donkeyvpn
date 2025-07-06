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
	cfg := config.LoadConfigFromEnvVars()

	e := echo.New()
	app, err := app.NewApplication(cfg, e)

	if err != nil {
		log.Error("error while creating a new DonkeyVPN application")
		os.Exit(1)
		return
	}
	app.Start()
}
