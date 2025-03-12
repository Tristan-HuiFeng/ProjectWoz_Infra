package main

// func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
// 	// Assuming discoveryRepo and AWS Config are provided from somewhere

// 	for _, message := range sqsEvent.Records {
// 		log.Info().Str("messageID", message.MessageId).Msg("Processing SQS message")

// 		var job Job
// 		err := json.Unmarshal([]byte(message.Body), &job)
// 		if err != nil {
// 			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Failed to unmarshal SQS message body")
// 			continue
// 		}

// 		// Run discovery with the parsed event data
// 		jobID, err := awscloud.RunDiscovery(cfg, discoveryRepo, resources)
// 		if err != nil {
// 			log.Error().Err(err).Str("messageID", message.MessageId).Msg("Error running discovery")
// 			continue
// 		}

// 		log.Info().Str("messageID", message.MessageId).Str("jobID", jobID.Hex()).Msg("Discovery process completed for message")
// 	}

// 	return nil
// }

// func main() {
// 	lambda.Start(handler)
// }
