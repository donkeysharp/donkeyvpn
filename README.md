Donkey VPN - Ephemeral VPNs
===

## Running Locally
Run the following commands
```
$ go get
$ source .env
$ go run cmd/bot/main.go
# This will run the application on port 8080
```

In order to test the Telegram bot by running it locally, you can create an Ngrok tunnel by running:

```
$ ngrok --url <free-ngrok-static-domain> http 8080
# This will generate an Ngrok url you can add to your testing Telegram bot as webhook
```

### Setting up your testing Telegram bot
After you created a Telegram Bot (for testing purposes)
Creating a webhook

https://cec0-2800-cd0-1274-c000-224e-7b16-80a0-8556.ngrok-free.app/telegram/donkeyvpn/webhook


payload.json
```json
{
  "url": "WEBHOOK_URL",
  "secret_token": "secret"
}
```

```sh
$ curl -H 'content-type: application/json' -XPOST --data @payload.json -sS https://api.telegram.org/bot${TELEGRAM_BOT_API_TOKEN}/setWebhook
```


## Resources
- https://www.wireguard.com/quickstart/
- https://dev.to/tangramvision/what-they-don-t-tell-you-about-setting-up-a-wireguard-vpn-1h2g
