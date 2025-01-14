package service

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/gommon/log"
)

var ErrFailedToCreateWireguardPeer = fmt.Errorf("failed to create wireguard peer")

func NewWireguardPeerService(table *aws.DynamoDB) *PeerService {
	return &PeerService{
		table: table,
	}
}

type PeerService struct {
	table *aws.DynamoDB
}

func (s *PeerService) Create(item *models.WireguardPeer) (bool, error) {
	created, err := s.table.CreateRecord(item)
	if err != nil {
		log.Errorf("Failed to create wireguard peer %v", err.Error())
		return false, err
	}
	if !created {
		log.Errorf("Failed to create wireguard peer record")
		return false, ErrFailedToCreateWireguardPeer
	}
	log.Infof("Wireguard peer created successfully: %v", item)
	return true, nil
}

func (s *PeerService) List() ([]models.WireguardPeer, error) {
	items, err := s.table.ListRecords()
	if err != nil {
		log.Errorf("Failed getting wireguard peers from dynamodb %v", err.Error())
		return nil, err
	}
	peers, err := models.DynamoItemsToWireguardPeers(items)
	if err != nil {
		log.Errorf("Failed to parse peer items %v", err.Error())
		return nil, err
	}
	return peers, nil
}

func (s *PeerService) Delete() {

}
