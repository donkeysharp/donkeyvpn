package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/labstack/gommon/log"
)

type VPNInstance struct {
	Hostname string `dynamodbav:"Hostname"`
	PublicIP string `dynamodbav:"PublicIP"`
}

func NewVPNInstance(hostname, publicIP string) *VPNInstance {
	return &VPNInstance{
		Hostname: hostname,
		PublicIP: publicIP,
	}
}

func (i *VPNInstance) ToItem() map[string]types.AttributeValue {
	log.Infof("Calling ToItem: Hostname %v PublicIP: %v", i.Hostname, i.PublicIP)
	return map[string]types.AttributeValue{
		"Hostname": &types.AttributeValueMemberS{Value: i.Hostname},
		"PublicIP": &types.AttributeValueMemberS{Value: i.PublicIP},
	}
}
func (i *VPNInstance) PrimaryKey() string {
	return i.Hostname
}

func (i *VPNInstance) RangeKey() string {
	return i.PublicIP
}

func DynamoItemsToVPNInstances(items []map[string]types.AttributeValue) ([]VPNInstance, error) {
	var instances []VPNInstance
	err := attributevalue.UnmarshalListOfMaps(items, &instances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal instances: %w", err)
	}
	return instances, nil
}
