package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewCreateProcessor(client *telegram.Client, asg *aws.AutoscalingGroup, table *aws.DynamoDB) CreateProcessor {
	return CreateProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
			asg:    asg,
			table:  table,
		},
	}
}

type CreateProcessor struct {
	ProcessorShared
}

func (p CreateProcessor) CreateVPN(update *telegram.Update) error {
	asg, err := p.asg.GetInfo()
	if err != nil {
		log.Error("Error while getting ASG information")
		return err
	}
	if *asg.DesiredCapacity > 0 {
		log.Warnf("ASG already has %d instances", *asg.DesiredCapacity)
		msg := "There is already an ephemeral VPN server configured, use /check command to verify"
		err = p.Client.SendMessage(msg, update.Message.Chat)
		if err != nil {
			log.Errorf("Error sending message to Telegram. msg=%s", msg)
		}
		return nil
	}

	desiredCapacity := 1
	err = p.asg.UpdateCapacity(int32(desiredCapacity))
	if err != nil {
		log.Errorf("Error while updating capacity of ASG to %d", desiredCapacity)
		msg := "Sorry, there was an error while creating the ephemeral VPN server, try again"
		err2 := p.Client.SendMessage(msg, update.Message.Chat)
		if err2 != nil {
			log.Errorf("Error sending message to Telegram. msg=%s", msg)
		}
		return err
	}
	msg := "Processing request... once the vpn server is ready, "
	msg += "you will be notified or use the /check command to get available ephemeral VPNs."
	err = p.Client.SendMessage(msg, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", msg)
	}
	return nil
}

func (p CreateProcessor) CreatePeer(ipAddress, publicKey string, update *telegram.Update) error {
	var peer models.WireguardPeer = models.WireguardPeer{
		IPAddress: ipAddress,
		PublicKey: publicKey,
	}
	created, err := p.table.CreateRecord(&peer)
	if err != nil {
		log.Errorf("Error while adding wireguard peer %v", err)
		message := "Error while adding Wireguard peer"
		err2 := p.Client.SendMessage(message, update.Message.Chat)
		if err2 != nil {
			log.Errorf("Error sending message to Telegram. msg=%s", message)
		}

		return err
	}

	if !created {
		message := "Wireguard peer could not be added."
		log.Warn(message)
		err := p.Client.SendMessage(message, update.Message.Chat)
		if err != nil {
			log.Errorf("Error sending message to Telegram. msg=%s", message)
		}
		return fmt.Errorf("error while creating Wireguard peer")
	}

	message := "Wireguard peer added successfully."
	log.Info(message)
	err = p.Client.SendMessage(message, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", message)
	}
	return nil
}

func (p CreateProcessor) Process(args []string, update *telegram.Update) error {
	log.Infof("Processing '/create' command with args %v for chat %d", args, update.Message.Chat.ChatId)

	usage := getUsage()
	if len(args) >= 1 && args[0] == "vpn" {
		return p.CreateVPN(update)
	}

	if len(args) >= 3 && args[0] == "peer" {
		ipAddress := args[1]
		publicKey := args[2]
		return p.CreatePeer(ipAddress, publicKey, update)
	}

	err := p.Client.SendMessage(usage, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", usage)
	}
	return nil
}
