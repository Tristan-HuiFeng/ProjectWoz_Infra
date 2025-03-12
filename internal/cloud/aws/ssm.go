package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/rs/zerolog/log"
)

func GetParam(paramName string, secure bool) (string, error) {
	cfg, err := GetRoleConfig()

	if err != nil {
		log.Fatal().Err(err).Str("function", "GetParam").Msg("unable to get account role config")
	}

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
