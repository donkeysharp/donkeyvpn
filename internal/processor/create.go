package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewCreateProcessor(client *telegram.Client, vpnSvc *service.VPNService) CreateProcessor {
	return CreateProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
		vpnSvc: vpnSvc,
	}
}

type CreateProcessor struct {
	ProcessorShared
	vpnSvc *service.VPNService
}

func (p CreateProcessor) sendMessage(msg string, update *telegram.Update) {
	err2 := p.Client.SendMessage(msg, update.Message.Chat)
	if err2 != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", msg)
	}
}

func (p CreateProcessor) CreateVPN(update *telegram.Update) error {
	result, err := p.vpnSvc.Create()
	if err != nil {
		log.Error("VPN instance creation failed")
		if err == service.ErrMaxCapacity {
			msg := "Maximum capacity reached, cannot create more instances."
			p.sendMessage(msg, update)
			return err
		} else {
			log.Errorf("Error while creating vpn instance: %v", err.Error())
			p.sendMessage("VPN instance creation failed", update)
			return err
		}
	}

	if !result {
		log.Error("Although no error was raised, the result of instance creation is false")
		p.sendMessage("VPN instance creation failed", update)
		return nil
	}

	msg := "Processing request... once the vpn server is ready, "
	msg += "you will be notified or use the /list vpn command to get available ephemeral VPNs."
	p.sendMessage(msg, update)
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
