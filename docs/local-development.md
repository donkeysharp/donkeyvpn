Local Development
===

If you want to contribute to DonkeyVPN, you will need a way test your local changes locally, due to DonkeyVPN's nature of interacting with different third party services such as Telegram (receiving messages via a webhook) and AWS for accessing DynamoDB tables and Autoscaling Groups, I will share the setup I use.

First you will need to install DonkeyVPN normally. Following the [installation](installation.md) steps.

## Exposing local service via a tunnel
Telegram needs to call a public URL via HTTPS to send messages, due to that we will need to expose our local-running services via a tunnel, in my case I use [Ngrok](https://ngrok.com/) but you can use any alternative that let's you expose a local service to the internet via HTTPS.

Ngrok's free version allows you to have a static domain name that is generated randomly, that way you will be able to always use the same public Ngrok domain. You will need to [install Ngrok's client](https://ngrok.com/downloads/) and setup a token. You will need to follow Ngork's [getting started](https://ngrok.com/docs/getting-started/) documentation.

The way I run it locally is by running the next command:

```
$ ngrok http --url=my-static-domain-name.ngrok-free.app 8080
```

In this case port `8080` was selected as DonkeyVPN uses that port locally.

## Running DonkeyVPN locally
After running the normal [installation](installation.md) steps, you will need to create a `.env` file based on the `.env.example` file.

You can start the application by running the following command:
```
$ go run cmd/bot/main.go

{"time":"2025-04-11T21:49:58.61093766-04:00","level":"INFO","prefix":"-","file":"application.go","line":"75","message":"Running as like executable application"}

   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.12.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on [::]:8080
```

## Updating Telegram bot to use Ngrok's url as webhook
Run the following script to reconfigure your Telegram's bot setting

```
TELEGRAM_BOT_API_TOKEN=<<insert you bot api token>>
BASE_URL=https://my-static-domain-name.ngrok-free.app
SECRET_TOKEN=<< insert your secret webhook token you generated >>
PAYLOAD="{
  \"url\": \"$BASE_URL/v1/api/telegram/donkeyvpn/webhook\",
  \"secret_token\": \"$SECRET_TOKEN\",
  \"drop_pending_updates\": true
}
"

curl -H 'content-type: application/json' -XPOST --data "$PAYLOAD" -sS https://api.telegram.org/bot$TELEGRAM_BOT_API_TOKEN/setWebhook
```

This way any message sent to your bot via Telegram will be sent via a Webhook to your local DonkeyVPN service.

## If you need to modify userdata to use local service
So far, any new message sent via Telegram goes to the local service, however there is another part that by default will point to the AWS API Gateway. The [userdata configured](../deploy/terraform/templates/userdata.tpl.sh) for the Autoscaling group. When a new VPN instance is created, it will call the registered base url which by default is to the AWS API Gateway. In order to use the local service you will need to modify `terraform.tfvars` file located at `deploy/terraform/` directory. And add the following:

```
...
testing_userdata_api_base_url="https://my-static-domain-name.ngrok-free.app"
```

And run the following command

```
$ bash deploy/scripts/install.sh terraform_apply
```

No need to run the `/tmp/donkeyvpn-webhook-register.sh` script.

Congratulations! Now you can start contributing to DonkeyVPN
