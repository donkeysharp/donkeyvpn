package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewCreateProcessor(client *telegram.Client, vpnSvc *service.VPNService, peerSvc *service.PeerService) CreateProcessor {
	return CreateProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
		peerSvc: peerSvc,
		vpnSvc:  vpnSvc,
	}
}

type CreateProcessor struct {
	ProcessorShared
	vpnSvc  *service.VPNService
	peerSvc *service.PeerService
}

func (p CreateProcessor) CreateVPN(update *telegram.Update) error {
	result, err := p.vpnSvc.Create(update.Message.Chat.ChatId)
	if err != nil {
		log.Error("VPN instance creation failed")
		if err == service.ErrMaxCapacity {
			msg := "Maximum capacity reached, cannot create more instances."
			p.SendMessage(msg, update)
			return err
		} else if err == service.ErrVPNInstanceCreating {
			msg := "There is an instance that currently is being created."
			msg += " Wait for it to finish before creating a new one."
			p.SendMessage(msg, update)
			return err
		} else {
			log.Errorf("Error while creating vpn instance: %v", err.Error())
			p.SendMessage("VPN instance creation failed", update)
			return err
		}
	}

	if !result {
		log.Error("Although no error was raised, the result of instance creation is false")
		p.SendMessage("VPN instance creation failed", update)
		return nil
	}

	msg := "Processing request... once the vpn server is ready, "
	msg += "you will be notified or use the /list vpn command to get available ephemeral VPNs."
	p.SendMessage(msg, update)
	return nil
}

func (p CreateProcessor) CreatePeer(ipAddress, publicKey, username string, update *telegram.Update) error {
	var peer *models.WireguardPeer = &models.WireguardPeer{
		IPAddress: ipAddress,
		PublicKey: publicKey,
		Username:  username,
	}
	created, err := p.peerSvc.Create(peer)
	if err != nil {
		log.Errorf("Failed to create wireguard peer %v", err.Error())
		if err == service.ErrInvalidWireguardKey {
			p.SendMessage("Invalid wireguard key format, please use a valid key and try again.", update)
			return err
		}
		if err == service.ErrInvalidIPAddress {
			p.SendMessage(
				fmt.Sprintf("Invalid IP address, it must be in the %v range", p.peerSvc.CidrRange),
				update,
			)
			return err
		}
		p.SendMessage("Error adding wireguard peer, please try again.", update)
		return err
	}

	if !created {
		log.Warnf("Wireguard peer could not be added, result was 'false'")
		p.SendMessage("Wireguard peer could not be added, please try again.", update)
	}

	log.Infof("Wireguard peer added successfully")
	p.SendMessage("Wireguard peer added successfully", update)

	return nil
}

func (p CreateProcessor) Process(args []string, update *telegram.Update) error {
	log.Infof("Processing '/create' command with args %v for chat %d", args, update.Message.Chat.ChatId)

	if len(args) >= 1 && args[0] == "vpn" {
		return p.CreateVPN(update)
	}

	if len(args) >= 3 && args[0] == "peer" {
		ipAddress := args[1]
		publicKey := args[2]
		username := update.Message.Chat.Username
		if username == "" {
			username = fmt.Sprintf("%v", update.Message.From.Id)
		}
		return p.CreatePeer(ipAddress, publicKey, username, update)
	}

	usage := getUsage()
	p.SendMessage(usage, update)
	return nil
}
