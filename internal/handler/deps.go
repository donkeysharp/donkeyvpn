package handler

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
)

// MessageSender delivers a message to a Telegram chat. *telegram.Client
// implements it; tests use testutil.RecordingSender.
type MessageSender interface {
	SendMessage(message string, chat *telegram.Chat) error
}

// VPNService is the part of *service.VPNService the handlers use.
type VPNService interface {
	ListPending() ([]models.VPNInstance, error)
	Update(item models.VPNInstance) (*models.VPNInstance, error)
}

// RecordLister lists every record of a DynamoDB table.
type RecordLister interface {
	ListRecords() ([]map[string]types.AttributeValue, error)
}

var (
	_ MessageSender = (*telegram.Client)(nil)
	_ VPNService    = (*service.VPNService)(nil)
	_ RecordLister  = (*aws.DynamoDB)(nil)
)
