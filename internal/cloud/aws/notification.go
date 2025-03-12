package awscloud

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/joho/godotenv"
)

func SendEmail() error {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("wozrole"))
	if err != nil {
		return err
	}

	client := sesv2.NewFromConfig(cfg)

	// Build the email message
	email := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(os.Getenv("EMAIL_SENDER")),
		Destination: &types.Destination{
			ToAddresses: []string{os.Getenv("EMAIL_RECIPIENT")},
		},
		// Message: &types.Message{
		// 	Subject: &types.Content{Data: aws.String(subject)},
		// 	Body: &types.Body{
		// 		Html: &types.Content{Data: aws.String(htmlBody)},
		// 		Text: &types.Content{Data: aws.String(textBody)},
		// 	},
		// },
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String("Test Email Subject"), // Subject of the email
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String("Hello, this is a test email sent using Amazon SESv2."),
					},
				},
			},
		},
	}

	// Send the email
	_, err = client.SendEmail(context.TODO(), email)
	if err != nil {
		log.Fatalf("unable to send email, %v", err)
	}

	log.Println("Email sent successfully!")
	return nil
}
