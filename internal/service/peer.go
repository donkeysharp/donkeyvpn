package service

import (
	"encoding/base64"
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/gommon/log"
)

var ErrFailedToCreateWireguardPeer = fmt.Errorf("failed to create wireguard peer")
var ErrInvalidWireguardKey = fmt.Errorf("invalid wireguard key format")

func NewWireguardPeerService(table *aws.DynamoDB) *PeerService {
	return &PeerService{
		table: table,
	}
}

func isValidKey(key string) bool {
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	return err == nil && len(decodedKey) == 32
}

type PeerService struct {
	table *aws.DynamoDB
}

func (s *PeerService) Create(item *models.WireguardPeer) (bool, error) {
	if !isValidKey(item.PublicKey) {
		return false, ErrInvalidWireguardKey
	}

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
