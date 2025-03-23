package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewSettingsProcessor(client *telegram.Client, ssm *aws.SSM, cfg *config.DonkeyVPNConfig) *SettingsProcessor {
	return &SettingsProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
		ssm:    ssm,
		config: cfg,
	}
}

type SettingsProcessor struct {
	ProcessorShared
	ssm    *aws.SSM
	config *config.DonkeyVPNConfig
}

func (p *SettingsProcessor) Process(args []string, update *telegram.Update) error {
	log.Info("Processing '/settings' command")

	value, err := p.ssm.GetParameter(p.config.PublicKeySSMParam, true)
	if err != nil {
		log.Errorf("Failed to retrieve ssm parameter %v", err.Error())
		p.SendMessage("Failed to retrieve settings. Try again please.", update)
		return err
	}

	log.Infof("Generate settings response")
	message := ""
	message += fmt.Sprintf("*VPN Public Key*: `%v\n`", value)
	message += fmt.Sprintf("*Wireguard IP Range*: `%v`\n", p.config.WireguardCidrRange)
	message += "\\-\\-\\-\\-\\-\n"

	p.SendMessage(message, update)

	return nil
}
