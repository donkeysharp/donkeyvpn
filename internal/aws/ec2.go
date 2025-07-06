package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/context"
)

type EC2 struct {
	cfg    aws.Config
	client *ec2.Client
	ctx    context.Context
}

func NewEC2(ctx context.Context) (*EC2, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Could not load aws default config")
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)

	return &EC2{
		cfg:    cfg,
		client: client,
		ctx:    ctx,
	}, nil
}

func (a *EC2) DescribeInstance(instanceId string) (*types.Instance, error) {
	output, err := a.client.DescribeInstances(a.ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceId},
	})
	if err != nil {
		log.Errorf("Failed to describe EC2 instance %v", err.Error())
		return nil, err
	}

	if len(output.Reservations) <= 0 {
		log.Warnf("Instance %v not found", instanceId)
		return nil, fmt.Errorf("no instance found with the given id")
	}

	instances := output.Reservations[0].Instances
	for _, instance := range instances {
		if *instance.InstanceId == instanceId {
			return &instance, nil
		}
	}
	log.Warnf("Instance %v not found", instanceId)
	return nil, fmt.Errorf("no instance found with the given id")
}
