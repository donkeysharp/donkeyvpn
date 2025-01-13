package handler

import (
	"fmt"
	"net/http"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type PeerHandler struct {
	WebhookSecret string
	PeersTable    *aws.DynamoDB
}

func (p *PeerHandler) List(c echo.Context) error {
	result, err := p.PeersTable.ListRecords()
	if err != nil {
		log.Errorf("error listing peers %v", err)
		return c.String(http.StatusInternalServerError, "Error listing peers")
	}

	peers, err := models.DynamoItemsToWireguardPeers(result)
	if err != nil {
		log.Errorf("error converting dynamodb items to wireguard peers: %v", err)
		return c.String(http.StatusInternalServerError, "error listing wireguard peers")
	}

	message := ""
	c.Response().Header().Add("content-type", "text/csv")
	for _, peer := range peers {
		message += fmt.Sprintf("%s,%s\n", peer.IPAddress, peer.PublicKey)
	}

	return c.String(http.StatusAccepted, message)
}
