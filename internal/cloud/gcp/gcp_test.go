package gcpcloud_test

import (
	"os"
	"testing"

	gcpcloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/gcp"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func TestGcsService(t *testing.T) {
	// Load environment variables
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}

	projectID := os.Getenv("GCP_PROJECT_ID")
	serviceAccount := os.Getenv("GCP_SERVICE_ACCOUNT")

	// Call ListBuckets with the loaded GCP config
	log.Info().Msg("Testing ListBuckets with real GCP credentials...")

	gcsService := &gcpcloud.GcsService{
		ProjectId:      projectID,
		ServiceAccount: serviceAccount,
	}

	buckets, err := gcsService.Discover()
	if err != nil {
		t.Errorf("Error discovering GCS buckets: %v", err)
	}

	log.Info().Msgf("Found %d buckets: %v", len(buckets), buckets)

	// Call getBucketPolicy with the loaded GCP config
	log.Info().Msg("Testing getBucketPolicy with real GCP credentials...")

	configs, err := gcsService.RetrieveConfig(buckets)
	if err != nil {
		t.Errorf("Error retrieving bucket policy: %v", err)
	}

	for bucketName, config := range configs {
		log.Info().Msgf("Bucket policy for %s: %v", bucketName, config["bucket_policy"])
		log.Info().Msgf("Bucket metadata for %s: %v", bucketName, config["bucket_metadata"])
	}
}
