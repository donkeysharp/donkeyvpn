package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

func (s *SSM) Exists(name string) bool {
	_, err := s.GetParameter(name, true)
	if err != nil {
		return false
	}
	return true
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

func (s *SSM) SetParameter(name string, value string, withEncryption, overwrite bool) (bool, error) {
	_, err := s.client.PutParameter(s.ctx, &ssm.PutParameterInput{
		Name:      aws.String(name),
		Value:     aws.String(value),
		Type:      types.ParameterTypeSecureString,
		Overwrite: aws.Bool(overwrite),
	})
	if err != nil {
		log.Errorf("Failed to set SSM parameter, error: %v", err.Error())
		return false, err
	}
	log.Infof("Parameter %v created successfully", name)
	return true, nil

}
