package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/labstack/gommon/log"
)

type WireguardPeer struct {
	IPAddress string `dynamodbav:"IPAddress"`
	PublicKey string `dynamodbav:"PublicKey"`
	Username  string `dynamodbav:"Username"`
}

func NewWireguardPeer(ipAddress, publicKey, username string) *WireguardPeer {
	return &WireguardPeer{
		IPAddress: ipAddress,
		PublicKey: publicKey,
		Username:  username,
	}
}

func (p *WireguardPeer) ToItem() map[string]types.AttributeValue {
	log.Infof("Calling ToItem: IPAddress %v PublicKey: %v", p.IPAddress, p.PublicKey)
	return map[string]types.AttributeValue{
		"IPAddress": &types.AttributeValueMemberS{Value: p.IPAddress},
		"PublicKey": &types.AttributeValueMemberS{Value: p.PublicKey},
		"Username":  &types.AttributeValueMemberS{Value: p.Username},
	}
}
func (p *WireguardPeer) PrimaryKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"IPAddress": &types.AttributeValueMemberS{Value: p.IPAddress},
	}
}

func (p *WireguardPeer) RangeKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"PublicKey": &types.AttributeValueMemberS{Value: p.PublicKey},
	}
}

func DynamoItemsToWireguardPeers(items []map[string]types.AttributeValue) ([]WireguardPeer, error) {
	var peers []WireguardPeer
	err := attributevalue.UnmarshalListOfMaps(items, &peers)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal peers: %w", err)
	}
	return peers, nil
}
