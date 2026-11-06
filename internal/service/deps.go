package service

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/donkeysharp/donkeyvpn/internal/aws"

	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// The interfaces below list only the methods the services call. The concrete
// clients in internal/aws satisfy them, and internal/testutil has in-memory
// fakes for unit tests.

type peerTable interface {
	CreateRecord(item aws.DynamoDBItem) (bool, error)
	GetRecord(item aws.DynamoDBItem) (map[string]types.AttributeValue, error)
	ListRecords() ([]map[string]types.AttributeValue, error)
	DeleteRecord(item aws.DynamoDBItem) error
}

type instanceTable interface {
	peerTable
	UpdateRecord(item aws.UpdateItem) error
	ListRecordsWithFilters(filter *aws.DynamoDBFilter) ([]map[string]types.AttributeValue, error)
}

type scalingGroup interface {
	GetInfo() (*asgtypes.AutoScalingGroup, error)
	UpdateCapacity(desiredCapacity int32) error
	DeleteInstance(instanceId string) error
}

type instanceDescriber interface {
	DescribeInstance(instanceId string) (*ec2types.Instance, error)
}

var (
	_ peerTable         = (*aws.DynamoDB)(nil)
	_ instanceTable     = (*aws.DynamoDB)(nil)
	_ scalingGroup      = (*aws.AutoscalingGroup)(nil)
	_ instanceDescriber = (*aws.EC2)(nil)
)
