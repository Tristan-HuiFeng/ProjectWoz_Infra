package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

var (
	discoveryRepo awscloud.DiscoveryRepository
	resources     []awscloud.ResourceDiscovery
)

type Job struct {
	ClientID string `json:"client_id"`
}

func init() {

	MONGO_DB_STRING, err := awscloud.GetParam(os.Getenv("MONGO_DB_STRING_PARAM"), true)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get db env from ssm")
	}
	os.Setenv("MONGO_DB_STRING", MONGO_DB_STRING)

	db, err := database.New()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to db")
	}
	defer db.Disconnect()

	discoveryRepo = awscloud.NewDiscoveryRepository(db)

	resources = []awscloud.ResourceDiscovery{
		&awscloud.S3Service{},
	}

}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	// Assuming discoveryRepo and AWS Config are provided from somewhere

	for _, message := range sqsEvent.Records {
		log.Info().Str("messageID", message.MessageId).Msg("Processing SQS message")

		var job Job
		err := json.Unmarshal([]byte(message.Body), &job)
		if err != nil {
			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Failed to unmarshal SQS message body")
			continue
		}

		// config, err := awscloud.ClientRoleConfig("arn:aws:iam::050752608470:role/WozCrossAccountRole")
		cfg, err := awscloud.ClientRoleConfig(fmt.Sprintf("arn:aws:iam::%s:role/WozCrossAccountRole", job.ClientID))
		if err != nil {
			log.Fatal().Msgf("unable to load SDK config, %v", err)
		}

		// Run discovery with the parsed event data
		jobID, err := awscloud.RunDiscovery(cfg, discoveryRepo, resources)
		if err != nil {
			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Error running discovery")
			continue
		}

		log.Info().Str("messageID", message.MessageId).Str("jobID", jobID.Hex()).Msg("Discovery process completed for message")
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
