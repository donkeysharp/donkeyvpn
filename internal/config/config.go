package config

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
