package service

import (
	"encoding/base64"
	"fmt"
	"net"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/gommon/log"
)

var ErrFailedToCreateWireguardPeer = fmt.Errorf("failed to create wireguard peer")
var ErrInvalidWireguardKey = fmt.Errorf("invalid wireguard key format")
var ErrInvalidIPAddress = fmt.Errorf("invalid IP address range")
var ErrWireguardPeerNotFound = fmt.Errorf("wireguard peer not found")

func NewWireguardPeerService(table *aws.DynamoDB, cidrRange string) *PeerService {
	return &PeerService{
		table:     table,
		CidrRange: cidrRange,
	}
}

func isValidKey(key string) bool {
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	return err == nil && len(decodedKey) == 32
}

func isValidIPAddress(ipAddress, cidrRange string) bool {
	_, cidrNet, err := net.ParseCIDR(cidrRange)
	if err != nil {
		return false
	}
	return cidrNet.Contains(net.ParseIP(ipAddress))
}

type PeerService struct {
	table     *aws.DynamoDB
	CidrRange string
}

func (s *PeerService) Create(item *models.WireguardPeer) (bool, error) {
	if !isValidKey(item.PublicKey) {
		return false, ErrInvalidWireguardKey
	}

	if !isValidIPAddress(item.IPAddress, s.CidrRange) {
		return false, ErrInvalidIPAddress
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

func (s *PeerService) Get(ipAddress string) (*models.WireguardPeer, error) {
	item, err := s.table.GetRecord(&models.WireguardPeer{IPAddress: ipAddress})

	if err != nil {
		log.Errorf("Failed to get wireguar peer with ip %v: %v", ipAddress, err.Error())
		return nil, err
	}
	if item == nil {
		log.Warnf("Wireguard peer with ip %v not found", ipAddress)
		return nil, ErrWireguardPeerNotFound
	}

	peer, err := models.DynamoItemToWireguardPeer(item)
	if err != nil {
		return nil, err
	}
	return peer, nil
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

func (s *PeerService) Delete(peerIP string) (bool, error) {
	peer, err := s.Get(peerIP)
	if err != nil {
		return false, err
	}

	err = s.table.DeleteRecord(peer)
	if err != nil {
		log.Errorf("Error deleting wireguard peer with ip %v: %v", peer.IPAddress, err.Error())
		return false, err
	}
	log.Infof("Wireguard peer with ip %v deleted successfully", peer.IPAddress)
	return true, nil
}
