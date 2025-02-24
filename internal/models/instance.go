package models

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/telegram"
	"github.com/labstack/gommon/log"
)

type VPNInstance struct {
	Id         string `dynamodbav:"Id"`
	Hostname   string `dynamodbav:"Hostname"`
	Port       string `dynamodbav:"Port"`
	Status     string `dynamodbav:"Status"`
	InstanceId string `dynamodbav:"InstanceId"`
	ChatId     string `dynamodbav:"ChatId"`
}

func NewVPNInstance(id, hostname, port, status, instanceId, chatId string) *VPNInstance {
	return &VPNInstance{
		Id:         id,
		Hostname:   hostname,
		Port:       port,
		Status:     status,
		InstanceId: instanceId,
		ChatId:     chatId,
	}
}

func (i VPNInstance) String() string {
	return fmt.Sprintf("ID: %v, Hostname: %v, Port: %v, Status: %v, InstanceId: %v, ChatId: %v", i.Id, i.Hostname, i.Port, i.Status, i.InstanceId, i.ChatId)
}

func (i *VPNInstance) ChatIdValue() telegram.ChatId {
	value, _ := strconv.ParseUint(i.ChatId, 10, 64)
	return telegram.ChatId(value)
}

func (i VPNInstance) ToItem() map[string]types.AttributeValue {
	log.Infof("Calling ToItem: Hostname %v Id: %v", i.Hostname, i.Id)
	return map[string]types.AttributeValue{
		"Id":         &types.AttributeValueMemberS{Value: i.Id},
		"Hostname":   &types.AttributeValueMemberS{Value: i.Hostname},
		"Port":       &types.AttributeValueMemberS{Value: i.Port},
		"Status":     &types.AttributeValueMemberS{Value: i.Status},
		"InstanceId": &types.AttributeValueMemberS{Value: i.InstanceId},
		"ChatId":     &types.AttributeValueMemberS{Value: i.ChatId},
	}
}
func (i VPNInstance) PrimaryKey() map[string]types.AttributeValue {
	log.Debugf("VPN Instance Primary Key: %v", i.Id)
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
	update := expression.Set(expression.Name("Hostname"), expression.Value(i.Hostname))
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

func FilterInstanceByStatus(status string) *aws.DynamoDBFilter {
	filterExpression := "#status = :status"
	// Status is a reserved keyword
	attributeNames := map[string]string{
		"#status": "Status",
	}
	attributeValues := map[string]types.AttributeValue{
		":status": &types.AttributeValueMemberS{Value: status},
	}
	return &aws.DynamoDBFilter{
		FilterExpression: &filterExpression,
		AttributeNames:   attributeNames,
		AttributeValues:  attributeValues,
	}
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
