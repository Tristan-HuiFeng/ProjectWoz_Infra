package gcpcloud

import (
	"context"
	"fmt"
	"time"

	cloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ConfigRepository interface {
	Create(job *ConfigRepository) error
	InsertMany(resourceConfigs []interface{}) ([]interface{}, error)
	FindByDiscoveryID(discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error)
	FindByTypeAndJobID(resourceType string, discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error)
	// FindByID(id bson.ObjectID) (*DiscoveryJob, error)
	// UpdateResources(id bson.ObjectID, resources map[string][]string) error
	// UpdateJob(id bson.ObjectID, resourceName string, resourceData []string) error
	// UpdateStatus(id bson.ObjectID, status string) error
}

type configRepository struct {
	collection *mongo.Collection
}

func (r *configRepository) Create(job *ConfigRepository) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//snapshot.ID := bson.NewObjectID()

	_, err := r.collection.InsertOne(ctx, job)
	return err
}

func NewConfigRepository(db database.Service) ConfigRepository {
	return &configRepository{
		collection: db.GetCollection("gcp_config"),
	}
}

func (r *configRepository) InsertMany(resourceConfigs []interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the documents into the MongoDB collection
	insertResult, err := r.collection.InsertMany(ctx, resourceConfigs)
	if err != nil {
		log.Error().Err(err).Str("function", "InsertMany").Msg("Failed to insert resource configurations")
		return nil, fmt.Errorf("failed to insert resource configurations: %w", err)
	}

	log.Info().Str("function", "InsertMany").
		Int("insertedCount", len(insertResult.InsertedIDs)).
		Msg("Resource configurations inserted successfully")

	// Return inserted IDs
	return insertResult.InsertedIDs, nil
}

func (r *configRepository) FindByDiscoveryID(discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error) {

	// Create a filter to match the discoveryJobID field
	filter := bson.M{
		"discovery_job_id": discoveryJobID,
	}

	// Find the documents that match the filter
	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []cloud.ResourceConfig
	for cursor.Next(context.Background()) {
		var config cloud.ResourceConfig
		if err := cursor.Decode(&config); err != nil {
			return nil, err
		}
		results = append(results, config)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *configRepository) FindByTypeAndJobID(resourceType string, discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error) {
	// Define the filter for matching resource_type and discovery_job_id
	filter := bson.M{
		"resource_type":    resourceType,
		"discovery_job_id": discoveryJobID,
	}

	// Use Find to get matching documents
	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %v", err)
	}
	defer cursor.Close(context.Background())

	// Store the results in a slice
	var results []cloud.ResourceConfig
	for cursor.Next(context.Background()) {
		var config cloud.ResourceConfig
		if err := cursor.Decode(&config); err != nil {
			log.Fatal().Err(err).Str("function", "FindByTypeAndJobID").Msg("issues with decoding db")
			return nil, err
		}
		results = append(results, config)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal().Err(err).Str("function", "FindByTypeAndJobID").Msg("failed to retrieve from db")
		return nil, err
	}

	log.Info().Str("function", "FindByTypeAndJobID").
		Int("discovery id", len(discoveryJobID)).
		Msg("retrieve resource configurations by discovery id and type successfully")

	return results, nil
}
