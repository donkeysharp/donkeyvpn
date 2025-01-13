package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/context"
)

type DynamoDB struct {
	TableName string
	client    *dynamodb.Client
	ctx       context.Context
}

type DynamoDBItem interface {
	ToItem() map[string]types.AttributeValue
	PrimaryKey() map[string]types.AttributeValue
	RangeKey() map[string]types.AttributeValue
}

type UpdateItem interface {
	DynamoDBItem
	UpdateExpression() (*expression.Expression, error)
}

func NewDynamoDB(ctx context.Context, tableName string) (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Could not load aws default config")
		return nil, err
	}
	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDB{
		TableName: tableName,
		client:    client,
		ctx:       ctx,
	}, nil
}

func (d *DynamoDB) CreateRecord(item DynamoDBItem) (bool, error) {
	_, err := d.client.PutItem(d.ctx, &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item:      item.ToItem(),
	})
	if err != nil {
		log.Errorf("Error while creating dynamodb record %v", err)
		return false, err
	}
	return true, nil
}

func (d *DynamoDB) UpdateRecord(item UpdateItem) error {
	expr, err := item.UpdateExpression()
	if err != nil {
		return err
	}

	output, err := d.client.UpdateItem(d.ctx, &dynamodb.UpdateItemInput{
		TableName:                 &d.TableName,
		Key:                       item.PrimaryKey(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})
	if err != nil {
		log.Errorf("Failed updating dynamodb record %v", err.Error())
		return err
	}

	log.Infof("Record updated successfully. Output metadata ", output.ResultMetadata)
	log.Infof("Record updated successfully. Output attributes ", output.Attributes)
	return nil
}

func (d *DynamoDB) GetRecord(item DynamoDBItem) (map[string]types.AttributeValue, error) {
	result, err := d.client.GetItem(d.ctx, &dynamodb.GetItemInput{
		TableName: &d.TableName,
		Key:       item.PrimaryKey(),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Item) == 0 {
		log.Warnf("There is no data for this record: %v", item)
		return nil, nil
	}
	return result.Item, nil
}

func (d *DynamoDB) ListRecords() ([]map[string]types.AttributeValue, error) {
	log.Info("Listing DynamoDB records")
	output, err := d.client.Scan(d.ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.TableName),
	})
	if err != nil {
		log.Errorf("Error listing records: %v", err)
		return nil, fmt.Errorf("failed to list peers: %w", err)
	}
	log.Info("Items listed successfully")

	return output.Items, nil
}

func (d *DynamoDB) DeleteRecord(item DynamoDBItem) error {
	_, err := d.client.DeleteItem(d.ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.TableName),
		Key:       item.ToItem(),
	})
	if err != nil {
		return fmt.Errorf("failed to delete peer: %w", err)
	}
	return nil
}
