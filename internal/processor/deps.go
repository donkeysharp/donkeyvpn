package processor

import (
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

// MessageSender delivers a reply to a Telegram chat. *telegram.Client
// implements it; tests use testutil.RecordingSender.
type MessageSender interface {
	SendMessage(message string, chat *telegram.Chat) error
}

// VPNService is the part of *service.VPNService the processors use.
type VPNService interface {
	Create(chatId telegram.ChatId) (bool, error)
	ListArray() ([]models.VPNInstance, error)
	Delete(vpnId string) (bool, error)
	ResetInstances() error
}

// PeerService is the part of *service.PeerService the processors use.
type PeerService interface {
	Create(peer *models.WireguardPeer) (bool, error)
	List() ([]models.WireguardPeer, error)
	Delete(peerIP string) (bool, error)
	CidrRange() string
}

// ParameterReader reads a value from SSM Parameter Store.
type ParameterReader interface {
	GetParameter(name string, withDecryption bool) (string, error)
}

var (
	_ MessageSender   = (*telegram.Client)(nil)
	_ VPNService      = (*service.VPNService)(nil)
	_ PeerService     = (*service.PeerService)(nil)
	_ ParameterReader = (*aws.SSM)(nil)
)
