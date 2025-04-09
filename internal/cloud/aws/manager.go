package awscloud

import (
	"fmt"
	"time"

	cloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/rs/zerolog/log"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const (
	InProgressStatus = "in-progress"
	CompletedStatus  = "completed"
)

type ResourceDiscovery interface {
	Discover(cfg aws.Config) ([]string, error)                                                      // Discover resources for a specific AWS service
	RetrieveConfig(cfg aws.Config, resourceIDs []string) (map[string]map[string]interface{}, error) // Retrieve resource configuration
	Name() string
}

func NewDiscoveryJob() *cloud.DiscoveryJob {
	return &cloud.DiscoveryJob{
		Status:    InProgressStatus,
		Resources: make(map[string][]string),
		CreatedAt: time.Now().Unix(),
	}
}

func RunDiscovery(cfg aws.Config, discoveryRepo DiscoveryRepository, clientID string, accountID string, resources []ResourceDiscovery) (bson.ObjectID, error) {
	log.Info().Msg("Starting discovery process...")

	job := NewDiscoveryJob()
	job.AccountID = accountID

	jobID, err := discoveryRepo.Create(job, clientID, accountID)
	if err != nil {
		log.Error().Err(err).Str("function", "RunDiscovery").Str("jobID", jobID.Hex()).Msg("Failed to create discovery job")
		return bson.NilObjectID, fmt.Errorf("RunDiscovery: %w", err)
	}
	log.Info().Str("jobID", jobID.Hex()).Msg("Discovery job created")

	for _, resource := range resources {
		resourceName := resource.Name()
		log.Info().Str("jobID", jobID.Hex()).Str("resource", resourceName).Msg("Starting resource discovery")

		resourceIDs, err := resource.Discover(cfg)

		if err != nil {
			log.Error().Err(err).Str("jobID", jobID.Hex()).Str("resource", resourceName).Msg("Failed to discover resources")
			return bson.NilObjectID, fmt.Errorf("RunDiscovery: %w", err)
		}

		err = discoveryRepo.UpdateJob(jobID, resourceName, resourceIDs)
		if err != nil {
			log.Error().Err(err).Str("function", "RunDiscovery").Str("jobID", jobID.Hex()).Str("resource", resourceName).Msg("Failed to update job with resources")
			return bson.NilObjectID, fmt.Errorf("RunDiscovery: %w", err)
		}

	}

	// s3ARNs, err := S3BucketList(cfg)
	// if err != nil {
	// 	log.Error().Err(err).Str("jobID", jobID.Hex()).Str("resourceType", S3ResourceType).Msg("Failed to retrieve S3 ARNs")

	// 	return bson.NilObjectID, fmt.Errorf("RunDiscovery: %w", err)
	// }

	// err = discoveryRepo.UpdateJob(jobID, S3ResourceType, s3ARNs)
	// if err != nil {
	// 	log.Error().Err(err).Str("function", "RunDiscovery").Str("jobID", jobID.Hex()).Str("resourceType", S3ResourceType).Msg("Failed to update job with resources")
	// 	return bson.NilObjectID, fmt.Errorf("failed to update job with S3 ARNs for job ID %s: %w", jobID.Hex(), err)
	// }

	err = discoveryRepo.UpdateStatus(jobID, CompletedStatus)
	if err != nil {
		log.Error().Err(err).Str("function", "RunDiscovery").Str("jobID", jobID.Hex()).Msg("Failed to update job status to complete")
		return bson.NilObjectID, fmt.Errorf("RunDiscovery: %w", err)
	}

	log.Info().Str("jobID", jobID.Hex()).Msg("Discovery process completed successfully")
	return jobID, nil
}

func RunRetrieval(cfg aws.Config, discoveryRepo DiscoveryRepository, configRepo ConfigRepository, discoveryID bson.ObjectID, clientID string, accountID string, resources []ResourceDiscovery) error {
	log.Info().Str("discoveryID", discoveryID.Hex()).Msg("Starting retrieval process...")

	discoveryJob, err := discoveryRepo.FindByID(discoveryID)
	if err != nil {
		log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Msg("Failed to find discovery job")
		return fmt.Errorf("retrival: %w", err)
	}

	var resourceConfigs []interface{}

	for _, resource := range resources {
		resourceName := resource.Name()
		log.Info().Str("discoveryID", discoveryID.Hex()).Str("resource", resourceName).Msg("Retrieving resource configurations")

		// Get discovered resource IDs for this resource type
		resourceIDs, found := discoveryJob.Resources[resourceName]
		if !found {
			log.Warn().Str("discoveryID", discoveryID.Hex()).Str("resource", resourceName).Msg("No discovered resources found for this type")
			continue
		}

		// Retrieve configurations for the discovered resource IDs
		configs, err := resource.RetrieveConfig(cfg, resourceIDs)
		if err != nil {
			log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resourceName).Msg("Failed to retrieve resource config")
			continue
		}

		log.Info().Str("discoveryID", discoveryID.Hex()).Str("resource", resourceName).Int("retrieved", len(configs)).Msg("Configurations retrieved successfully")

		for resourceID, config := range configs {
			resourceConfig := cloud.ResourceConfig{
				DiscoveryJobID: discoveryID,
				ClientID:       clientID,
				AccountID:      accountID,
				Provider:       "AWS",
				ResourceType:   resourceName,
				ResourceID:     resourceID,
				Config:         config,
			}
			resourceConfigs = append(resourceConfigs, resourceConfig)
		}

	}

	// configs, err := S3ConfigRetrival(cfg, discoveryJob.Resources[S3ResourceType])
	// if err != nil {
	// 	log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resourceType", S3ResourceType).Msg("Failed to retrieve S3 config")
	// 	return fmt.Errorf("Retrival: %w", err)
	// }

	// for bucketName, policy := range configs {
	// 	resourceConfig := AWSResourceConfig{
	// 		DiscoveryJobID: discoveryID,
	// 		ResourceType:   "s3",
	// 		ResourceID:     bucketName,
	// 		Config:         policy,
	// 	}
	// 	resourceConfigs = append(resourceConfigs, resourceConfig)
	// }

	// result, err := configRepo.InsertMany(resourceConfigs)
	// if err != nil {
	// 	log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Msg("Failed to insert resource configs for S3")
	// 	return fmt.Errorf("Retrival: %w", err)
	// }

	// 	log.Info().Str("discoveryID", discoveryID.Hex()).Int("inserted", len(result)).Msg("Configurations inserted successfully")

	if len(resourceConfigs) > 0 {
		result, err := configRepo.InsertMany(resourceConfigs)
		if err != nil {
			log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Msg("Failed to insert resource configs")
			return fmt.Errorf("RunRetrival: %w", err)
		}
		log.Info().Str("discoveryID", discoveryID.Hex()).Int("inserted", len(result)).Msg("Configurations inserted successfully")
	} else {
		log.Warn().Str("discoveryID", discoveryID.Hex()).Msg("No configurations retrieved for any resources")
	}

	return nil
}
