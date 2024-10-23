Donkey VPN - Ephemeral VPNs
===


## Setting up your own ephemeral bot agent

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
