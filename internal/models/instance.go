package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/labstack/gommon/log"
)

type VPNInstance struct {
	Id         string `dynamodbav:"Id"`
	Hostname   string `dynamodbav:"Hostname"`
	Port       string `dynamodbav:"Port"`
	Status     string `dynamodbav:"Status"`
	InstanceId string `dynamodbav:"InstanceId"`
}

func NewVPNInstance(id, hostname, port, status, instanceId string) *VPNInstance {
	return &VPNInstance{
		Id:         id,
		Hostname:   hostname,
		Port:       port,
		Status:     status,
		InstanceId: instanceId,
	}
}

func (i VPNInstance) ToItem() map[string]types.AttributeValue {
	log.Infof("Calling ToItem: Hostname %v Id: %v", i.Hostname, i.Id)
	return map[string]types.AttributeValue{
		"Id":         &types.AttributeValueMemberS{Value: i.Id},
		"Hostname":   &types.AttributeValueMemberS{Value: i.Hostname},
		"Port":       &types.AttributeValueMemberS{Value: i.Port},
		"Status":     &types.AttributeValueMemberS{Value: i.Status},
		"InstanceId": &types.AttributeValueMemberS{Value: i.InstanceId},
	}
}
func (i VPNInstance) PrimaryKey() map[string]types.AttributeValue {
	log.Infof("VPN Instance Primary Key: %v", i.Id)
	return map[string]types.AttributeValue{
		"Id": &types.AttributeValueMemberS{Value: i.Id},
	}
}

func (i VPNInstance) RangeKey() map[string]types.AttributeValue {
	log.Infof("VPN Instance Range key: %v", i.Hostname)
	return map[string]types.AttributeValue{
		"Hostname": &types.AttributeValueMemberS{Value: i.Hostname},
	}
}

func (i VPNInstance) UpdateExpression() (*expression.Expression, error) {
	update := expression.Set(expression.Name("Hostname"), expression.Value(i.Id))
	update.Set(expression.Name("Port"), expression.Value(i.Port))
	update.Set(expression.Name("Status"), expression.Value(i.Status))
	update.Set(expression.Name("InstanceId"), expression.Value(i.InstanceId))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Errorf("Failed to create update expression: %v", err.Error())
		return nil, err
	}
	return &expr, err
}

func DynamoItemToVPNInstance(item map[string]types.AttributeValue) (*VPNInstance, error) {
	var instance VPNInstance
	err := attributevalue.UnmarshalMap(item, &instance)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal instance: %v", err.Error())
	}
	return &instance, nil
}

func DynamoItemsToVPNInstances(items []map[string]types.AttributeValue) ([]VPNInstance, error) {
	var instances []VPNInstance
	err := attributevalue.UnmarshalListOfMaps(items, &instances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal instances: %w", err)
	}
	return instances, nil
}

func DynamoItemsToVPNInstancesMap(items []map[string]types.AttributeValue) (map[string]VPNInstance, error) {
	var instanceMap map[string]VPNInstance = make(map[string]VPNInstance)
	var instanceList []VPNInstance
	err := attributevalue.UnmarshalListOfMaps(items, &instanceList)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal instances: %w", err)
	}

	for _, instance := range instanceList {
		instanceMap[instance.Id] = instance
	}
	return instanceMap, nil
}
