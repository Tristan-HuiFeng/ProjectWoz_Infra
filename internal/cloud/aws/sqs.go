package awscloud

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog/log"
)

func SendSQSMessage(message string, client *sqs.Client, queueURL string) error {

	sendMessageOutput, err := client.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(message),
	})

	if err != nil {
		log.Fatal().Err(err).Str("function", "SendSQSMessage").Str("queueURL", queueURL).Msg("unable to send message")
		return fmt.Errorf("unable to send message: %w", err)
	}

	log.Info().Err(err).Str("function", "SendSQSMessage").Str("queueURL", queueURL).Msg(*sendMessageOutput.MessageId)

	return nil

}
