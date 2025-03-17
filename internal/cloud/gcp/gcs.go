package gcpcloud

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

type GcsService struct {
	ProjectId string
	ServiceAccount string
}

func (s *GcsService) Name() string {
	return "gcs"
}

func (s *GcsService) Discover() ([]string, error) {
	ctx := context.Background()
	return listBuckets(ctx, s.ProjectId, s.ServiceAccount)
}

func (s *GcsService) RetrieveConfig(bucketNames []string) (map[string]map[string]interface{}, error) {
	configs := make(map[string]map[string]interface{})

	for _, bucket := range bucketNames {
		if configs[bucket] == nil {
			configs[bucket] = make(map[string]interface{})
		}

		policy, err := getBucketPolicy(bucket, s.ServiceAccount)

		if err != nil {
			continue // Skip to the next bucket
		}

		configs[bucket]["bucket_policy"] = policy
	}

	return configs, nil
}

// Sets up GCS Client using STS token + service account impersonation
func setupGcsClient(ctx context.Context, serviceAccount string) (*storage.Service, error) {
	ts, err := ImpersonateServiceAccount(serviceAccount)

	if err != nil {
		return nil, fmt.Errorf("failed to impersonate service account: %v", err)
	}

	storageService, err := storage.NewService(ctx, option.WithTokenSource(*ts))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %v", err)
	}

	log.Info().Msg("Created storage service")
	
	return storageService, nil
}


// ListBuckets lists all buckets in the GCP project
func listBuckets(ctx context.Context, projectID string, serviceAccount string) ([]string, error) {
	service, err := setupGcsClient(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Listing buckets in project %s", projectID)
	
	buckets, err := service.Buckets.List(projectID).Do()
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Found %d buckets", len(buckets.Items))
	
	var bucketNames []string
	for _, bucket := range buckets.Items {
		bucketNames = append(bucketNames, bucket.Name)
	}
	
	return bucketNames, nil
}

// GetBucketPolicy gets the bucket IAM policy.
func getBucketPolicy(bucketName string, serviceAccount string) (*storage.Policy, error) {
	ctx := context.Background()
	service, err := setupGcsClient(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Getting IAM policy for bucket %s", bucketName)

	policy, err := service.Buckets.GetIamPolicy(bucketName).Do()
	if err != nil {
		return nil, err
	}

	return policy, nil
}
