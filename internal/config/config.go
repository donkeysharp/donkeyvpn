package config

import "os"

type DonkeyVPNConfig struct {
	TelegramBotAPIToken  string
	WebhookSecret        string
	AutoscalingGroupName string
	PeersTableName       string
	InstancesTableName   string
	PublicKeySSMParam    string
	WireguardCidrRange   string
	RunAsLambda          bool
}

func LoadConfigFromEnvVars() DonkeyVPNConfig {
	var telegramBotAPIToken string = os.Getenv("TELEGRAM_BOT_API_TOKEN")
	var webhookSecret string = os.Getenv("WEBHOOK_SECRET")
	var autoscalingGroupName string = os.Getenv("AUTOSCALING_GROUP_NAME")
	var peersTableName string = os.Getenv("DYNAMODB_PEERS_TABLE_NAME")
	var instancesTableName string = os.Getenv("DYNAMODB_INSTANCES_TABLE_NAME")
	var runAsLambdaStr string = os.Getenv("RUN_AS_LAMBDA")
	var publicKeySSMParam string = os.Getenv("SSM_PUBLIC_KEY")
	var wireguardCidrRange string = os.Getenv("WIREGUARD_CIDR_RANGE")
	var runAsLambda bool = false
	if runAsLambdaStr == "true" {
		runAsLambda = true
	}

	return DonkeyVPNConfig{
		TelegramBotAPIToken:  telegramBotAPIToken,
		WebhookSecret:        webhookSecret,
		AutoscalingGroupName: autoscalingGroupName,
		PeersTableName:       peersTableName,
		InstancesTableName:   instancesTableName,
		PublicKeySSMParam:    publicKeySSMParam,
		WireguardCidrRange:   wireguardCidrRange,
		RunAsLambda:          runAsLambda,
	}
}
