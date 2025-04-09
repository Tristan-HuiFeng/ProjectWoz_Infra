package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/opa2"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	regoRepo          opa2.RegoRepository
	scanRepo          opa2.ScanRepository
	configRepo        awscloud.ConfigRepository
	awsResources      []awscloud.ResourceDiscovery
	client            database.Service
	sqsClient         *sqs.Client
	processingRoleCfg aws.Config
)

type Message struct {
	ClientID    string `json:"client_id"`
	JobID       string `json:"job_id"`
	ClientEmail string `json:"client_email"`
	Provider    string `json:"provider"`
}

func init() {

	log.Info().Str("function", "init").Msg("getting db param")

	processingRoleCfg, err := awscloud.GetRoleConfig()
	if err != nil {
		log.Fatal().Err(err).Str("function", "init").Msg("unable to get account role config")
	}

	// processing role env
	processingRole, err := awscloud.GetParam(os.Getenv("PROCESSING_ROLE"), false, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get processing role env from ssm")
	}
	os.Setenv("PROCESSING_ROLE", processingRole)

	// db setup
	MONGO_DB_STRING, err := awscloud.GetParam(os.Getenv("MONGO_DB_STRING_PARAM"), true, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get db env from ssm")
	}
	os.Setenv("MONGO_DB_STRING", MONGO_DB_STRING)

	log.Info().Str("function", "init").Msg("setting db conn")
	client, err := database.New()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to db")
	}

	var c = make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	go func() {

		sig := <-c
		log.Info().Str("signal", sig.String()).Msg("Signal received, shutting down")
		client.Disconnect()

	}()

	SMTP_PASSWORD, err := awscloud.GetParam(os.Getenv("SMTP_PASSWORD_PARAM"), true, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get db env from ssm")
	}
	os.Setenv("SMTP_PASSWORD", SMTP_PASSWORD)

	// SCAN_QUEUE_URL, err := awscloud.GetParam(os.Getenv("SCAN_QUEUE_URL"), false, processingRoleCfg)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("unable to get scan queue url from ssm")
	// }
	// os.Setenv("SCAN_QUEUE_URL", SCAN_QUEUE_URL)

	regoRepo = opa2.NewRegoRepository(client)
	scanRepo = opa2.NewScanRepository(client)
	configRepo = awscloud.NewConfigRepository(client)

	awsResources = []awscloud.ResourceDiscovery{
		&awscloud.S3Service{},
	}

	sqsClient = sqs.NewFromConfig(processingRoleCfg)

}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	// Assuming discoveryRepo and AWS Config are provided from somewhere

	for _, message := range sqsEvent.Records {
		log.Info().Str("messageID", message.MessageId).Msg("Processing SQS message")

		var job Message
		err := json.Unmarshal([]byte(message.Body), &job)
		if err != nil {
			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Failed to unmarshal SQS message body")
			continue
		}

		id, err := bson.ObjectIDFromHex(job.JobID)
		if err != nil {
			log.Fatal().Msgf("unable to convert job id to bson.ObjectID, %v", err)
		}

		if job.Provider == "AWS" {
			err = opa2.RunScan(configRepo, scanRepo, regoRepo, id, awsResources, job.ClientEmail, job.ClientID)
		} else if job.Provider == "GCP" {
			log.Info().Msg("WIP")
		} else {
			log.Warn().Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("Provider not supported")
			return errors.New("provider not supported")
		}

		if err != nil {
			log.Fatal().Msgf("Scan failed, %v", err)
		}

		log.Info().Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("Scan process completed for message")
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
