package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	gcpcloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/gcp"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

var (
	awsDiscoveryRepo awscloud.DiscoveryRepository
	gcpDiscoveryRepo gcpcloud.DiscoveryRepository
	awsResources     []awscloud.ResourceDiscovery
	gcpResources     []gcpcloud.ResourceDiscovery
	sqsClient        *sqs.Client
	// processingRoleCfg aws.Config
	// client        database.Service
)

// type DiscoveryJob struct {
// 	ClientID string `json:"client_id"`
// }

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

	RETRIEVAL_QUEUE_URL, err := awscloud.GetParam(os.Getenv("RETRIEVAL_QUEUE_PARAM"), false, processingRoleCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get retrival queue url from ssm")
	}
	os.Setenv("RETRIEVAL_QUEUE_URL", RETRIEVAL_QUEUE_URL)

	awsDiscoveryRepo = awscloud.NewDiscoveryRepository(client)
	gcpDiscoveryRepo = gcpcloud.NewDiscoveryRepository(client)

	awsResources = []awscloud.ResourceDiscovery{
		&awscloud.S3Service{},
	}

	gcpResources = []gcpcloud.ResourceDiscovery{
		&gcpcloud.GcsService{},
	}

	sqsClient = sqs.NewFromConfig(processingRoleCfg)

}

func awsHandler(clientID string, accountID string, clientEmail string) error {
	log.Info().Str("account id", accountID).Msg("setting up discovery for aws client")
	// config, err := awscloud.ClientRoleConfig("arn:aws:iam::050752608470:role/WozCrossAccountRole")
	cfg, err := awscloud.ClientRoleConfig(fmt.Sprintf("arn:aws:iam::%s:role/WozCrossAccountRole", accountID))
	if err != nil {
		log.Fatal().Msgf("unable to load SDK config, %v", err)
		return err
	}

	// Run discovery with the parsed event data
	jobID, err := awscloud.RunDiscovery(cfg, awsDiscoveryRepo, clientID, accountID, awsResources)
	if err != nil {
		log.Error().Err(err).Str("account id", accountID).Msg("Error running discovery")
		return err
	}

	msg := Message{
		JobID:       jobID.Hex(),
		ClientID:    clientID,
		AccountID:   accountID,
		ClientEmail: clientEmail,
		Provider:    "AWS",
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

	log.Info().Str("account id", accountID).Str("jobID", jobID.Hex()).Msg("discovery process completed for aws client")

	return nil
}

func gcpHandler(clientID string, projectID, clientEmail string) error {

	log.Info().Str("project id", projectID).Msg("setting up discovery for gcp client")

	jobID, err := gcpcloud.RunDiscovery(gcpDiscoveryRepo, clientID, projectID, gcpResources)
	if err != nil {
		log.Error().Err(err).Str("project id", projectID).Msg("Error running discovery")
		return err
	}

	log.Info().Str("project id", projectID).Str("jobID", jobID.Hex()).Msg("discovery process completed for gcp client")

	msg := Message{
		JobID:       jobID.Hex(),
		ClientID:    clientID,
		AccountID:   projectID,
		ClientEmail: clientEmail,
		Provider:    "GCP",
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

	return nil

	// ctx := context.Background()
	// client, err := storage.NewClient(ctx)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to create client for gcp storage")
	// 	return err
	// }
	// defer client.Close()

	// ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	// defer cancel()

	// it := client.Buckets(ctx, clientGCPProjectID)
	// for {
	// 	battrs, err := it.Next()
	// 	if err == iterator.Done {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatal().Err(err).Msg("failed while iterating gcp buckets")
	// 		return err
	// 	}
	// 	log.Info().Msgf("Bucket: %v\n", battrs.Name)
	// }
	// return nil
}

type Invoke struct {
	InvokeType   string `json:"invoke_type"`
	ClientID     string `json:"client_id"`
	AwsAccountID string `json:"aws_account_id"`
	GcpProjectID string `json:"gcp_project_id"`
	ClientEmail  string `json:"client_email"`
}

func handler(ctx context.Context, event json.RawMessage) error {
	// Assuming discoveryRepo and AWS Config are provided from somewhere

	log.Info().Msg("running interval discovery")

	var e events.LambdaFunctionURLRequest
	if err := json.Unmarshal(event, &e); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return err
	}

	fmt.Println(e.Body)

	var invoke Invoke
	if err := json.Unmarshal([]byte(e.Body), &invoke); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return err
	}

	log.Info().Str("invoke type", invoke.InvokeType).Str("aws acc id", invoke.AwsAccountID).Str("gcp proj id", invoke.GcpProjectID).Msg("invoke debug")

	if invoke.InvokeType == "manual" {

		log.Info().Msg("manual trigger invoked")
	} else {
		log.Info().Msg("interval trigger invoked")
	}

	// 	if invoke.AwsAccountID != "" {
	// 		awsHandler(invoke.ClientID, invoke.AwsAccountID, invoke.ClientEmail)
	// 	}

	// 	if invoke.GcpProjectID != "" {
	// 		gcpHandler(invoke.ClientID, invoke.GcpProjectID, invoke.ClientEmail)
	// 	}

	// 	log.Info().Msg("manual discovery process completed")

	// } else {

	// 	log.Info().Msg("interval trigger invoked")
	// 	awsClientID := "1"
	// 	awsAccountID := "050752608470"
	// 	clientEmail := "user.ad.proj@gmail.com"
	// 	gcpClientID := "1"
	// 	clientGCPProjectID := "the-other-450607-a4"
	// 	// clientGCPProjectID := "cs464-454011"

	// 	awsHandler(awsClientID, awsAccountID, clientEmail)

	// 	gcpHandler(gcpClientID, clientGCPProjectID, clientEmail)

	// 	log.Info().Msg("interval discovery process completed")

	// }

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
