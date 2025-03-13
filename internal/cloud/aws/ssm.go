package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetParam(paramName string, secure bool, cfg aws.Config) (string, error) {

	ssmClient := ssm.NewFromConfig(cfg)

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(secure),
	}

	paramOutput, err := ssmClient.GetParameter(context.Background(), paramInput)
	if err != nil {
		return "", fmt.Errorf("failed to get parameter: %w", err)
	}

	return *paramOutput.Parameter.Value, nil
}
