package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	gcpcloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/gcp"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	awsDiscoveryRepo awscloud.DiscoveryRepository
	awsConfigRepo    awscloud.ConfigRepository
	awsResources     []awscloud.ResourceDiscovery

	gcpDiscoveryRepo gcpcloud.DiscoveryRepository
	gcpConfigRepo    gcpcloud.ConfigRepository
	gcpResources     []gcpcloud.ResourceDiscovery

	client            database.Service
	sqsClient         *sqs.Client
	processingRoleCfg aws.Config
)

type Message struct {
	JobID       string `json:"job_id"`
	ClientID    string `json:"client_id"`
	AccountID   string `json:"account_id"`
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

	SCAN_QUEUE_URL, err := awscloud.GetParam(os.Getenv("SCAN_QUEUE_PARAM"), false, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get scan queue url from ssm")
	}
	os.Setenv("SCAN_QUEUE_URL", SCAN_QUEUE_URL)

	awsDiscoveryRepo = awscloud.NewDiscoveryRepository(client)
	awsConfigRepo = awscloud.NewConfigRepository(client)

	awsResources = []awscloud.ResourceDiscovery{
		&awscloud.S3Service{},
	}

	gcpDiscoveryRepo = gcpcloud.NewDiscoveryRepository(client)
	gcpConfigRepo = gcpcloud.NewConfigRepository(client)

	gcpResources = []gcpcloud.ResourceDiscovery{
		&gcpcloud.GcsService{},
	}

	sqsClient = sqs.NewFromConfig(processingRoleCfg)

}

func awsHandler(job Message) error {
	cfg, err := awscloud.ClientRoleConfig(fmt.Sprintf("arn:aws:iam::%s:role/WozCrossAccountRole", job.AccountID))
	if err != nil {
		log.Fatal().Msgf("unable to load SDK config, %v", err)
	}

	id, err := bson.ObjectIDFromHex(job.JobID)
	if err != nil {
		log.Fatal().Msgf("unable to convert job id to bson.ObjectID, %v", err)
	}

	err = awscloud.RunRetrieval(cfg, awsDiscoveryRepo, awsConfigRepo, id, job.ClientID, job.AccountID, awsResources)
	if err != nil {
		log.Fatal().Msgf("Retrieval failed for aws, %v", err)
	}

	return nil
}

func gcpHandler(job Message) error {

	id, err := bson.ObjectIDFromHex(job.JobID)
	if err != nil {
		log.Fatal().Msgf("unable to convert job id to bson.ObjectID, %v", err)
	}

	//err = gcpcloud.RunRetrieval(discoveryRepo, configRepo, id, job.ClientID, job.AccountID, resources)
	err = gcpcloud.RunRetrival(gcpDiscoveryRepo, gcpConfigRepo, id, job.ClientID, job.AccountID, gcpResources)
	if err != nil {
		log.Fatal().Msgf("Retrieval failed for gcp, %v", err)
	}

	return nil
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

		if job.Provider == "AWS" {
			log.Info().Str("account_id", job.AccountID).Str("messageID", message.MessageId).Msg("retrieving config for aws message")
			err = awsHandler(job)
			if err != nil {
				log.Fatal().Err(err).Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("AWS Handler Error")
				return errors.New("error running AWS Handler")
			}
		} else if job.Provider == "GCP" {
			log.Info().Str("account_id", job.AccountID).Str("messageID", message.MessageId).Msg("retrieving config for gcp message")
			err = gcpHandler(job)
			if err != nil {
				log.Fatal().Err(err).Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("AWS Handler Error")
				return errors.New("error running GCP Handler")
			}
		} else {
			log.Warn().Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("Provider not supported")
			return errors.New("provider not supported")
		}

		err = awscloud.SendSQSMessage(string(message.Body), sqsClient, os.Getenv("SCAN_QUEUE_URL"))
		if err != nil {
			log.Fatal().Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("Failed to send message to scan queue")
		}

		log.Info().Str("messageID", message.MessageId).Str("jobID", job.JobID).Msg("Retrieval process completed for message")
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
