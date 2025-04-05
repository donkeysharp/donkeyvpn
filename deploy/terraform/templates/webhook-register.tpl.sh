#!/bin/bash

TELEGRAM_BOT_API_TOKEN=${IN_TELEGRAM_BOT_API_TOKEN}
BASE_URL=${IN_BASE_URL}
SECRET_TOKEN=${IN_SECRET_TOKEN}
PAYLOAD="{
  \"url\": \"$BASE_URL/v1/api/telegram/donkeyvpn/webhook\",
  \"secret_token\": \"$SECRET_TOKEN\",
  \"drop_pending_updates\": true
}
"

curl -H 'content-type: application/json' -XPOST --data "$PAYLOAD" -sS https://api.telegram.org/bot$TELEGRAM_BOT_API_TOKEN/setWebhook
