package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

var (
	discoveryRepo awscloud.DiscoveryRepository
	resources     []awscloud.ResourceDiscovery
	sqsClient     *sqs.Client
	// processingRoleCfg aws.Config
	// client        database.Service
)

// type DiscoveryJob struct {
// 	ClientID string `json:"client_id"`
// }

type Message struct {
	ClientID    string `json:"client_id"`
	JobID       string `json:"job_id"`
	ClientEmail string `json:"client_email"`
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

	RETRIEVAL_QUEUE_URL, err := awscloud.GetParam(os.Getenv("RETRIEVAL_QUEUE_PARAM"), false, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get retrival queue url from ssm")
	}
	os.Setenv("RETRIEVAL_QUEUE_URL", RETRIEVAL_QUEUE_URL)

	discoveryRepo = awscloud.NewDiscoveryRepository(client)

	resources = []awscloud.ResourceDiscovery{
		&awscloud.S3Service{},
	}

	sqsClient = sqs.NewFromConfig(processingRoleCfg)

}

func awsHandler(clientID string, clientEmail string) error {
	log.Info().Str("client id", clientID).Msg("setting up discovery for aws client")
	// config, err := awscloud.ClientRoleConfig("arn:aws:iam::050752608470:role/WozCrossAccountRole")
	cfg, err := awscloud.ClientRoleConfig(fmt.Sprintf("arn:aws:iam::%s:role/WozCrossAccountRole", clientID))
	if err != nil {
		log.Fatal().Msgf("unable to load SDK config, %v", err)
		return err
	}

	// Run discovery with the parsed event data
	jobID, err := awscloud.RunDiscovery(cfg, discoveryRepo, clientID, resources)
	if err != nil {
		log.Error().Err(err).Str("client id", clientID).Msg("Error running discovery")
		return err
	}

	msg := Message{
		ClientID:    clientID,
		JobID:       jobID.Hex(),
		ClientEmail: clientEmail,
	}

	messageBody, err := json.Marshal(msg)
	if err != nil {
		log.Fatal().Msgf("failed to marshal message into JSON, %v", err)
		return err
	}

	err = awscloud.SendSQSMessage(string(messageBody), sqsClient, os.Getenv("RETRIEVAL_QUEUE_URL"))
	if err != nil {
		log.Fatal().Str("jobID", jobID.Hex()).Msg("failed to send message to retrieval queue")
		return err
	}

	log.Info().Str("client id", clientID).Str("jobID", jobID.Hex()).Msg("discovery process completed for aws client")
}

func handler(ctx context.Context) error {
	// Assuming discoveryRepo and AWS Config are provided from somewhere

	log.Info().Msg("running interval discovery")

	clientID := "050752608470"
	clientEmail := "user.ad.proj@gmail.com"

	awsHandler(clientID, clientEmail)

	log.Info().Msg("interval discovery process completed")

	return nil

}

// func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
// 	// Assuming discoveryRepo and AWS Config are provided from somewhere

// 	for _, message := range sqsEvent.Records {
// 		log.Info().Str("messageID", message.MessageId).Msg("Processing SQS message")

// 		var discoveryJob DiscoveryJob
// 		err := json.Unmarshal([]byte(message.Body), &discoveryJob)
// 		if err != nil {
// 			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Failed to unmarshal SQS message body")
// 			continue
// 		}

// 		// config, err := awscloud.ClientRoleConfig("arn:aws:iam::050752608470:role/WozCrossAccountRole")
// 		cfg, err := awscloud.ClientRoleConfig(fmt.Sprintf("arn:aws:iam::%s:role/WozCrossAccountRole", discoveryJob.ClientID))
// 		if err != nil {
// 			log.Fatal().Msgf("unable to load SDK config, %v", err)
// 		}

// 		// Run discovery with the parsed event data
// 		jobID, err := awscloud.RunDiscovery(cfg, discoveryRepo, resources)
// 		if err != nil {
// 			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Error running discovery")
// 			continue
// 		}

// 		msg := Message{
// 			ClientID: discoveryJob.ClientID,
// 			JobID:    jobID.Hex(),
// 		}

// 		messageBody, err := json.Marshal(msg)
// 		if err != nil {
// 			log.Fatal().Msgf("failed to marshal message into JSON, %v", err)
// 		}

// 		err = awscloud.SendSQSMessage(string(messageBody), sqsClient, os.Getenv("RETRIEVAL_QUEUE_URL"))
// 		if err != nil {
// 			log.Fatal().Str("messageID", message.MessageId).Str("jobID", jobID.Hex()).Msg("Failed to send message to retrieval queue")
// 		}

// 		log.Info().Str("messageID", message.MessageId).Str("jobID", jobID.Hex()).Msg("Discovery process completed for message")
// 	}

// 	return nil
// }

func main() {
	lambda.Start(handler)
}
