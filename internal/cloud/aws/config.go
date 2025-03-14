package awscloud

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/rs/zerolog/log"
)

func ClientRoleConfig(clientRoleARN string) (aws.Config, error) {
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("wozrole"))
	log.Info().Str("function", "GetRoleConfig").Msg("retriving aws role config")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatal().Err(err).Str("function", "ClientRoleConfig").Msg("failed to load default config for aws")
		return aws.Config{}, err
	}

	// assuming our own role with permission for cross account
	appCreds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), os.Getenv("PROCESSING_ROLE"))

	_, err = appCreds.Retrieve(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Str("function", "ClientRoleConfig").Msg("failed to retrieve aws credentials for woz processing role")
		return aws.Config{}, err
	}

	cfg.Credentials = aws.NewCredentialsCache(appCreds)

	// assuming client role
	appCreds = stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), clientRoleARN)

	_, err = appCreds.Retrieve(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Str("function", "ClientRoleConfig").Msg("failed to retrieve aws credentials for client role")
		return aws.Config{}, err
	}

	cfg.Credentials = aws.NewCredentialsCache(appCreds)

	log.Info().Str("function", "GetRoleConfig").Msg("retrived role config successfully")

	return cfg, nil

}

func GetRoleConfig() (aws.Config, error) {
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("wozrole"))
	log.Info().Str("function", "GetRoleConfig").Msg("retriving aws role config")
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatal().Err(err).Str("function", "GetRoleConfig").Msg("failed to load default config for aws")
		return aws.Config{}, err
	}

	log.Info().Str("function", "GetRoleConfig").Msg("retrived role config successfully")

	return cfg, nil
}
