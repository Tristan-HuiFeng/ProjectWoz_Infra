package awscloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/rs/zerolog/log"
)

type S3Service struct {
}

func (s *S3Service) Name() string {
	return "s3"
}

func (d *S3Service) Discover(cfg aws.Config) ([]string, error) {
	// func S3BucketList(client *s3.Client) ([]string, error) {
	client := s3.NewFromConfig(cfg)

	output, err := client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Error().Err(err).Msg("Error discovering s3 resource")
		return nil, fmt.Errorf("failed to list S3 buckets: %v", err)
	}

	// var bucketARNs []string
	// for _, bucket := range output.Buckets {
	// 	bucketArn := fmt.Sprintf("arn:aws:s3:::%s", *bucket.Name)
	// 	bucketARNs = append(bucketARNs, bucketArn)
	// }

	var bucketNames []string
	for _, bucket := range output.Buckets {
		bucketNames = append(bucketNames, *bucket.Name)
	}

	return bucketNames, nil
}

func (d *S3Service) RetrieveConfig(cfg aws.Config, bucketNames []string) (map[string]map[string]interface{}, error) {
	//func S3ConfigRetrival(client *s3.Client, bucketNames []string) (map[string]interface{}, error) {

	client := s3.NewFromConfig(cfg)

	configs := make(map[string]map[string]interface{})

	for _, bucket := range bucketNames {

		if configs[bucket] == nil {
			configs[bucket] = make(map[string]interface{})
		}

		output, err := client.GetBucketPolicy(context.TODO(), &s3.GetBucketPolicyInput{
			Bucket: aws.String(bucket),
		})

		if err != nil {
			// Check if the error is "NoSuchBucketPolicy" error (direct check)
			var oe *smithy.OperationError

			if errors.As(err, &oe) {
				//log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
				log.Info().Err(err).Str("bucket name:", bucket)
				configs[bucket]["bucket_policy"] = ""
			} else {
				log.Error().Err(err).Str("bucket name:", bucket).Msg("Error Retriving Bucket Policy")
			}

			continue // Skip to the next bucket
		}
		var unmarshalledPolicy map[string]interface{}
		err = json.Unmarshal([]byte(*output.Policy), &unmarshalledPolicy)
		if err != nil {
			log.Error().Err(err).Str("bucket name:", bucket).Msg("Error unmarshalling bucket policy")
			continue // Skip to the next bucket
		}

		configs[bucket]["bucket_policy"] = unmarshalledPolicy

	}

	return configs, nil
}
