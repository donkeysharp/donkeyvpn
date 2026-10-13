package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

type DonkeyVPNConfig struct {
	TelegramBotAPIToken    string
	WebhookSecret          string
	AutoscalingGroupName   string
	PeersTableName         string
	InstancesTableName     string
	PublicKeySSMParam      string
	WireguardCidrRange     string
	AllowedChatIdsSSMParam string
	RunAsLambda            bool
}

func ParseAllowedChatIds(raw string) []int64 {
	result := make([]int64, 0)
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		value, err := strconv.ParseInt(entry, 10, 64)
		if err != nil {
			log.Warnf("Ignoring invalid allowed chat id %q", entry)
			continue
		}
		result = append(result, value)
	}
	return result
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
	var allowedChatIdsSSMParam string = os.Getenv("SSM_ALLOWED_CHAT_IDS")
	var runAsLambda bool = false
	if runAsLambdaStr == "true" {
		runAsLambda = true
	}

	return DonkeyVPNConfig{
		TelegramBotAPIToken:    telegramBotAPIToken,
		WebhookSecret:          webhookSecret,
		AutoscalingGroupName:   autoscalingGroupName,
		PeersTableName:         peersTableName,
		InstancesTableName:     instancesTableName,
		PublicKeySSMParam:      publicKeySSMParam,
		WireguardCidrRange:     wireguardCidrRange,
		AllowedChatIdsSSMParam: allowedChatIdsSSMParam,
		RunAsLambda:            runAsLambda,
	}
}
