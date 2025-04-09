package gcpcloud

import (
	"context"
	"time"

	"cloud.google.com/go/storage"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

type GcsService struct {
	ProjectId      string
	ServiceAccount string
}

func (s *GcsService) Name() string {
	return "gcs"
}

func (s *GcsService) Discover(projectID string) ([]string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client for gcp storage")
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	var bucketNames []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal().Str("project id", projectID).Err(err).Msg("failed while iterating gcp buckets")
			return nil, err
		}
		bucketNames = append(bucketNames, battrs.Name)
		// log.Info().Msgf("Bucket: %v\n", battrs.Name)
	}

	return bucketNames, nil

	//return listBuckets(ctx, s.ProjectId, s.ServiceAccount)
}

func (s *GcsService) RetrieveConfig(bucketNames []string) (map[string]map[string]interface{}, error) {
	configs := make(map[string]map[string]interface{})

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client for gcp storage")
		return nil, err
	}
	defer client.Close()

	for _, bucket := range bucketNames {
		if configs[bucket] == nil {
			configs[bucket] = make(map[string]interface{})
		}

		bk := client.Bucket(bucket)

		policy, err := bk.IAM().Policy(ctx)
		if err != nil {
			log.Error().Err(err).Str("bucket name:", bucket).Msg("Error getting gcs policy")
			continue // Skip to the next bucket
		}

		// var unmarshalledPolicy map[string]interface{}
		// err = json.Unmarshal([]byte(*policy), &unmarshalledPolicy)
		// if err != nil {
		// 	log.Error().Err(err).Str("bucket name:", bucket).Msg("Error unmarshalling bucket policy")
		// 	continue // Skip to the next bucket
		// }

		configs[bucket]["bucket_policy"] = policy.InternalProto.Bindings

		metadata, err := bk.Attrs(ctx)
		if err != nil {
			log.Error().Err(err).Str("bucket name:", bucket).Msg("Error getting gcs metadata")
			continue // Skip to the next bucket
		}

		configs[bucket]["bucket_metadata"] = metadata

	}

	return configs, nil
}

// Sets up GCS Client using STS token + service account impersonation
// func setupGcsClient(ctx context.Context, serviceAccount string) (*storage.Service, error) {
// 	ts, err := ImpersonateServiceAccount(serviceAccount)

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to impersonate service account: %v", err)
// 	}

// 	storageService, err := storage.NewService(ctx, option.WithTokenSource(*ts))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create storage service: %v", err)
// 	}

// 	log.Info().Msg("Created storage service")

// 	return storageService, nil
// }

// ListBuckets lists all buckets in the GCP project
// func listBuckets(ctx context.Context, projectID string, serviceAccount string) ([]string, error) {
// 	service, err := setupGcsClient(ctx, serviceAccount)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Info().Msgf("Listing buckets in project %s", projectID)

// 	buckets, err := service.Buckets.List(projectID).Do()
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Info().Msgf("Found %d buckets", len(buckets.Items))

// 	var bucketNames []string
// 	for _, bucket := range buckets.Items {
// 		bucketNames = append(bucketNames, bucket.Name)
// 	}

// 	return bucketNames, nil
// }

// GetBucketPolicy gets the bucket IAM policy.
// func getBucketPolicy(bucketName string, serviceAccount string) (*storage.Policy, error) {
// 	ctx := context.Background()
// 	service, err := setupGcsClient(ctx, serviceAccount)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Info().Msgf("Getting IAM policy for bucket %s", bucketName)

// 	policy, err := service.Buckets.GetIamPolicy(bucketName).Do()
// 	if err != nil {
// 		return nil, err
// 	}

// 	return policy, nil
// }

// // getBucketMetadata gets the bucket metadata.
// func getBucketMetadata(bucketName string, serviceAccount string) (*storage.Bucket, error) {
// 	ctx := context.Background()
// 	service, err := setupGcsClient(ctx, serviceAccount)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Info().Msgf("Getting metadata for bucket %s", bucketName)

// 	bucket, err := service.Buckets.Get(bucketName).Do()

// 	if err != nil {
// 		return nil, fmt.Errorf("Bucket(%q).Attrs: %w", bucketName, err)
// 	}

// 	return bucket, nil
// }
