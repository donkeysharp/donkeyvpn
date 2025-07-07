package app

import (
	"context"
	"fmt"
	"time"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/config"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type DonkeyVPNNotifierApp struct {
	VPNService     *service.VPNService
	TelegramClient *telegram.Client
}

func NewNotifierApplication(cfg config.DonkeyVPNConfig, tgClient *telegram.Client) (*DonkeyVPNNotifierApp, error) {
	ctx := context.Background()
	ec2Client, err := aws.NewEC2(ctx)
	if err != nil {
		return nil, err
	}

	instancesTable, err := aws.NewDynamoDB(ctx, cfg.InstancesTableName)
	if err != nil {
		return nil, err
	}

	// asg client is not required by this application
	vpnService := service.NewVPNService(nil, instancesTable, ec2Client)

	return &DonkeyVPNNotifierApp{
		VPNService:     vpnService,
		TelegramClient: tgClient,
	}, nil
}

func (n *DonkeyVPNNotifierApp) CheckInstances() error {
	instances, err := n.VPNService.ListOlderThan(time.Hour * 1)
	if err != nil {
		log.Errorf("Failed to list instances older than <time>: %v", err.Error())
		return err
	}
	if len(instances) == 0 {
		log.Infof("No instances are running for more than 1 hour")
		return nil
	}

	message := "⚠️ Warning! The next instances have been running for *more than 1 hour:*\n\n"

	receivers := make(map[telegram.ChatId]struct{})

	for _, instance := range instances {
		log.Infof("Instance %v is running for more than an hour", instance.Id)
		receivers[instance.ChatIdValue()] = struct{}{}
		message += fmt.Sprintf("🟢 Instance ID: `%v`\n", instance.Id)
	}
	message += "----\n"
	message += "In case your are not using them, you can delete them and create new ones later 🤑"
	for receiver := range receivers {
		log.Infof("Sending notification to chat %v", receiver)
		err := n.TelegramClient.SendMessage(message, &telegram.Chat{
			ChatId: receiver,
		})
		if err != nil {
			log.Errorf("Failed to send telegram message: %v", err.Error())
		}
	}
	return nil
}
