package awscloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrJobNotFound = errors.New("discovery job not found")

type DiscoveryRepository interface {
	Create(job *cloud.DiscoveryJob, clientID string, accountID string) (bson.ObjectID, error)
	FindByID(id bson.ObjectID) (*cloud.DiscoveryJob, error)
	UpdateResources(id bson.ObjectID, resources map[string][]string) error
	UpdateJob(id bson.ObjectID, resourceName string, resourceData []string) error
	UpdateStatus(id bson.ObjectID, status string) error
}

type discoveryRepository struct {
	collection *mongo.Collection
}

func NewDiscoveryRepository(db database.Service) DiscoveryRepository {
	return &discoveryRepository{
		collection: db.GetCollection("aws_discovery"),
	}
}

func (r *discoveryRepository) Create(job *cloud.DiscoveryJob, clientID string, accountID string) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	job.ID = bson.NewObjectID()
	job.ClientID = clientID
	job.AccountID = accountID
	job.Provider = "AWS"

	result, err := r.collection.InsertOne(ctx, job)
	if err != nil {
		log.Error().Err(err).Str("function", "Create").Str("jobID", job.ID.Hex()).Msg("Failed to create new discovery job")
		return bson.NilObjectID, fmt.Errorf("failed to insert ObjectID %s: %w", job.ID.Hex(), err)
	}

	insertedID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		log.Error().Str("function", "Create").Str("jobID", job.ID.Hex()).Msg("Failed to convert inserted ID to ObjectID")
		return bson.NilObjectID, fmt.Errorf("failed to convert inserted ID to ObjectID %s", job.ID.Hex())
	}

	log.Info().Str("function", "Create").Str("jobID", insertedID.Hex()).Msg("Discovery job created successfully")
	return insertedID, nil
}

func (r *discoveryRepository) FindByID(id bson.ObjectID) (*cloud.DiscoveryJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var job cloud.DiscoveryJob
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Error().Err(err).Str("function", "FindByID").Str("jobID", id.Hex()).Msg("Discovery job not found")
			return nil, ErrJobNotFound
		}
		log.Error().Err(err).Str("function", "FindByID").Str("jobID", id.Hex()).Msg("Failed to find discovery job")
		return nil, fmt.Errorf("failed to find discovery job with ID %s: %w", id.Hex(), err)
	}

	log.Info().Str("function", "FindByID").Str("jobID", id.Hex()).Msg("Discovery job found successfully")
	return &job, nil
}

func (r *discoveryRepository) UpdateResources(id bson.ObjectID, resources map[string][]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"resources": resources}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("function", "UpdateResources").Str("jobID", id.Hex()).Msg("Failed to update resources for discovery job")
		return fmt.Errorf("failed to update resources for discovery job with ID %s: %w", id.Hex(), err)
	}

	log.Debug().Str("function", "UpdateResources").Str("jobID", id.Hex()).Int64("matchedCount", result.MatchedCount).Int64("modifiedCount", result.ModifiedCount).Msg("Update result")

	if result.MatchedCount == 0 && result.ModifiedCount == 0 {
		log.Warn().Str("function", "UpdateResources").Str("jobID", id.Hex()).Msg("No job found to update resources")
		return ErrJobNotFound
	}

	log.Info().Str("function", "UpdateResources").Str("jobID", id.Hex()).Msg("Resources updated successfully")
	return nil
}

func (r *discoveryRepository) UpdateJob(id bson.ObjectID, resourceName string, resourceData []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare update query
	update := bson.M{
		"$set": bson.M{
			"resources." + resourceName: resourceData,
		},
	}

	log.Debug().Str("function", "UpdateJob").Str("jobID", id.Hex()).Str("resourceName", resourceName).Msg("Updating job with new resources")

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("function", "UpdateJob").Str("jobID", id.Hex()).Str("resourceName", resourceName).Msg("Failed to update job")
		return fmt.Errorf("failed to update resources for discovery job with ID %s: %w", id.Hex(), err)
	}

	log.Debug().Str("function", "UpdateJob").Str("jobID", id.Hex()).Int64("matchedCount", result.MatchedCount).Int64("modifiedCount", result.ModifiedCount).Msg("Update result")

	if result.MatchedCount == 0 {
		log.Warn().Str("function", "UpdateJob").Str("jobID", id.Hex()).Str("resourceName", resourceName).Msg("No job found to update")
		return ErrJobNotFound
	}

	log.Info().Str("function", "UpdateJob").Str("jobID", id.Hex()).Str("resourceName", resourceName).Msg("Job updated with new resources successfully")
	return nil
}

func (r *discoveryRepository) UpdateStatus(id bson.ObjectID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prepare update query
	update := bson.M{"$set": bson.M{"status": status}}

	log.Debug().Str("function", "UpdateStatus").Str("jobID", id.Hex()).Str("status", status).Msg("Updating job status")

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("function", "UpdateStatus").Str("jobID", id.Hex()).Str("status", status).Msg("Failed to update job status")
		return fmt.Errorf("failed to update status with ID %s: %w", id.Hex(), err)
	}

	log.Debug().Str("function", "UpdateStatus").Str("jobID", id.Hex()).Int64("matchedCount", result.MatchedCount).Msg("Update result")

	if result.MatchedCount == 0 {
		log.Warn().Str("function", "UpdateStatus").Str("jobID", id.Hex()).Msg("No job found to update status")
		return ErrJobNotFound
	}

	log.Info().Str("function", "UpdateStatus").Str("jobID", id.Hex()).Str("status", status).Msg("Job status updated successfully")
	return nil
}
