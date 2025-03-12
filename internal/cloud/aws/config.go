package awscloud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/rs/zerolog/log"
)

func ClientRoleConfig(clientRoleARN string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("wozrole"))
	if err != nil {
		return aws.Config{}, err
	}

	appCreds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), clientRoleARN)

	_, err = appCreds.Retrieve(context.TODO())
	if err != nil {
		return aws.Config{}, err
	}

	cfg.Credentials = aws.NewCredentialsCache(appCreds)

	return cfg, nil

}

func GetRoleConfig() (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("wozrole"))
	if err != nil {
		log.Fatal().Err(err).Str("function", "GetRoleConfig").Msg("failed to retrieve role config")
		return aws.Config{}, err
	}

	return cfg, nil
}
