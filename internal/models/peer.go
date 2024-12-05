package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/labstack/gommon/log"
)

type WireguardPeer struct {
	IPAddress string `dynamodbav:"PeerAddress"`
	PublicKey string `dynamodbav:"PublicKey"`
}

func NewWireguardPeer(ipAddress, publicKey string) *WireguardPeer {
	return &WireguardPeer{
		IPAddress: ipAddress,
		PublicKey: publicKey,
	}
}

func (p *WireguardPeer) ToItem() map[string]types.AttributeValue {
	log.Infof("Calling ToItem: PeerAddress %v PublicKey: %v", p.IPAddress, p.PublicKey)
	return map[string]types.AttributeValue{
		"PeerAddress": &types.AttributeValueMemberS{Value: p.IPAddress},
		"PublicKey":   &types.AttributeValueMemberS{Value: p.PublicKey},
	}
}
func (p *WireguardPeer) PrimaryKey() string {
	return p.IPAddress
}

func (p *WireguardPeer) RangeKey() string {
	return p.PublicKey
}

func DynamoItemsToWireguardPeers(items []map[string]types.AttributeValue) ([]WireguardPeer, error) {
	var peers []WireguardPeer
	err := attributevalue.UnmarshalListOfMaps(items, &peers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal peers: %w", err)
	}
	return peers, nil
}
