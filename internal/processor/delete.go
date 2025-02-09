package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewDeleteProcessor(client *telegram.Client, vpnSvc *service.VPNService, peerSvc *service.PeerService) DeleteProcessor {
	return DeleteProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
		vpnSvc:  vpnSvc,
		peerSvc: peerSvc,
	}
}

type DeleteProcessor struct {
	ProcessorShared
	vpnSvc  *service.VPNService
	peerSvc *service.PeerService
}

func (p DeleteProcessor) sendMessage(msg string, update *telegram.Update) {
	err2 := p.Client.SendMessage(msg, update.Message.Chat)
	if err2 != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", msg)
	}
}

func (p DeleteProcessor) Process(args []string, update *telegram.Update) error {
	log.Infof("Processing '/delete' command with args %v for chat %d", args, update.Message.Chat.ChatId)

	if len(args) == 2 && args[0] == "vpn" {
		vpnId := args[1]
		result, err := p.vpnSvc.Delete(vpnId)
		if err != nil {
			p.sendMessage("Error while deleting VPN instance, please try again.", update)
			return err
		}
		if !result {
			log.Errorf("Although there was not error during deletion of vpn instance %v, the result was false", vpnId)
			p.sendMessage("Could not delete VPN instance.", update)
			return nil
		}
		p.sendMessage(fmt.Sprintf("VPN intance %v deleted successfully", vpnId), update)
		return nil
	}

	if len(args) == 2 && args[0] == "peer" {
		peerIP := args[1]
		result, err := p.peerSvc.Delete(peerIP)
		if err != nil {
			log.Errorf("Error deleting peer %v. Error: %v", peerIP, err)
			p.sendMessage("Error deleting VPN peer.", update)
			return err
		}

		if !result {
			log.Errorf("Although code didn't failed, it was not possible to delete peer %v", peerIP)
			p.sendMessage("Could not delete VPN peer.", update)
			return nil
		}

		p.sendMessage("Peer deleted successfully", update)
		return nil
	}

	usage := getUsage()
	err := p.Client.SendMessage(usage, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", usage)
	}
	return nil
}
