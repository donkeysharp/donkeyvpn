package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/labstack/gommon/log"
)

type SSM struct {
	client *ssm.Client
	ctx    context.Context
}

func NewSSM(ctx context.Context) (*SSM, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("Could not load aws default config for ssm client")
		return nil, err
	}
	client := ssm.NewFromConfig(cfg)

	return &SSM{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *SSM) GetParameter(name string, withDecryption bool) (string, error) {
	log.Infof("Retrieving SSM parameter %v", name)
	result, err := s.client.GetParameter(s.ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(withDecryption),
	})
	if err != nil {
		log.Errorf("Failed to retrieve SSM parameter, %v, error: %v", name, err.Error())
		return "", err
	}
	log.Infof("SSM parameter retrieved successfully")
	value := result.Parameter.Value
	return *value, nil
}
