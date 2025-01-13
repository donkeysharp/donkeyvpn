package processor

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

func NewListProcessor(client *telegram.Client, peersTable, instancesTable *aws.DynamoDB) ListProcessor {
	return ListProcessor{
		ProcessorShared: ProcessorShared{
			Client: client,
		},
		peersTable:     peersTable,
		instancesTable: instancesTable,
	}
}

type ListProcessor struct {
	ProcessorShared
	peersTable     *aws.DynamoDB
	instancesTable *aws.DynamoDB
}

func (p ListProcessor) ListVPNs(update *telegram.Update) error {
	log.Info("Listing vpn instances for telegram")
	result, err := p.instancesTable.ListRecords()
	if err != nil {
		log.Errorf("Error listing vpn instances: %v", err)
		return err
	}
	instances, err := models.DynamoItemsToVPNInstances(result)
	if err != nil {
		log.Errorf("Error converting dynamodb items to vpn instances: %v", err)
		return err
	}

	message := "List of instances:\n-----\n"
	for _, item := range instances {
		log.Infof("Proccessing instance: %s %s %s %s", item.Id, item.Hostname, item.Port, item.Status)
		message += fmt.Sprintf("Id: %s\n", item.Id)
		message += fmt.Sprintf("Hostname: %s\n", item.Hostname)
		message += fmt.Sprintf("Port: %s\n", item.Port)
		message += fmt.Sprintf("Status: %s\n", item.Status)
		message += "-----\n"
	}

	err = p.Client.SendMessage(message, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", message)
	}
	return nil
}

func (p ListProcessor) ListPeers(update *telegram.Update) error {
	log.Info("Listing peers for telegram")
	result, err := p.peersTable.ListRecords()
	if err != nil {
		log.Errorf("Error while listing peers: %v", err)
		return err
	}
	peers, err := models.DynamoItemsToWireguardPeers(result)
	if err != nil {
		log.Errorf("Error converting dynamodb items to wireguard peers: %v", err)
		return err
	}

	message := "List of peers:\n-----\n"
	for _, item := range peers {
		log.Infof("Processing peer: %s", item.IPAddress)
		message += fmt.Sprintf("IP Address: %s\n", item.IPAddress)
		message += fmt.Sprintf("Public Key: %s\n", item.PublicKey)
		message += "-----\n"
	}

	err = p.Client.SendMessage(message, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", message)
	}
	return nil
}

func (p ListProcessor) Process(args []string, update *telegram.Update) error {
	log.Infof("Processing '/list' command with args %v for chat %d", args, update.Message.Chat.ChatId)

	usage := getUsage()
	if len(args) >= 1 && args[0] == "vpn" {
		return p.ListVPNs(update)
	}

	if len(args) >= 1 && args[0] == "peers" {
		return p.ListPeers(update)
	}

	err := p.Client.SendMessage(usage, update.Message.Chat)
	if err != nil {
		log.Errorf("Error sending message to Telegram. msg=%s", usage)
	}
	return nil
}
