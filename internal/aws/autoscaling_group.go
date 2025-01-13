package aws

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/labstack/gommon/log"
	"golang.org/x/net/context"
)

type AutoscalingGroup struct {
	Name   string
	cfg    aws.Config
	client *autoscaling.Client
	ctx    context.Context
}

func NewAutoscalingGroup(ctx context.Context, name string) (*AutoscalingGroup, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Could not load aws default config")
		return nil, err
	}

	client := autoscaling.NewFromConfig(cfg)

	return &AutoscalingGroup{
		Name:   name,
		cfg:    cfg,
		client: client,
		ctx:    ctx,
	}, nil
}

func (a *AutoscalingGroup) GetInfo() (*types.AutoScalingGroup, error) {
	log.Info("Getting information from autoscaling group")
	res, err := a.client.DescribeAutoScalingGroups(a.ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{a.Name},
	})

	if err != nil {
		log.Error("Error while describing autoscaling group")
		return nil, err
	}

	if len(res.AutoScalingGroups) == 0 {
		log.Error("could not find autoscaling group")
		return nil, errors.New("could not find autoscaling group")
	}

	return &res.AutoScalingGroups[0], nil
}

func (a *AutoscalingGroup) UpdateCapacity(desiredCapacity int32) error {
	asg, err := a.GetInfo()
	if err != nil {
		return err
	}

	log.Infof("Updating autoscaling gorup capacity to desired capacity: %d", desiredCapacity)
	res, err := a.client.UpdateAutoScalingGroup(a.ctx, &autoscaling.UpdateAutoScalingGroupInput{
		DesiredCapacity:      &desiredCapacity,
		AutoScalingGroupName: &a.Name,
	})

	if err != nil {
		log.Errorf("Failed during ASG update from %d to %d. Error %v", *asg.DesiredCapacity, desiredCapacity, err)
		return err
	}

	log.Infof("ResultMetadata: %v", res.ResultMetadata)
	return nil
}

// func (a *AutoscalingGroup) GetIntances() error {
// 	asg, err := a.GetInfo()
// 	if err != nil {
// 		return err
// 	}

// 	if len(asg.Instances) > 0 {

// 	}
// 	/*asg.Instances[0].InstanceId
// 	for all instances get the instance from instance_id
// 	*/
// 	return nil

// }
