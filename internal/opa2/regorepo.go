package opa2

import (
	"context"
	"fmt"
	"time"

	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RegoRepository interface {
	Create(rego *RegoPolicy) (bson.ObjectID, error)
	FindByResourceType(resourceType string) (*RegoPolicy, error)
}

type regoRepository struct {
	collection *mongo.Collection
}

func NewRegoRepository(db database.Service) RegoRepository {
	return &regoRepository{
		collection: db.GetCollection("rego"),
	}
}

func (r *regoRepository) Create(rego *RegoPolicy) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingRego, _ := r.FindByResourceType(rego.ResourceType)
	if existingRego != nil {
		log.Error().Str("function", "Create").Str("ResourceType", rego.ResourceType).Msg("rego policy with the same resource type already exists")
		return bson.NilObjectID, fmt.Errorf("rego policy with resource type %s already exists", rego.ResourceType)
	}

	result, err := r.collection.InsertOne(ctx, rego)
	if err != nil {
		log.Error().Err(err).Str("function", "Create").Str("Rego ID", rego.ID.Hex()).Msg("failed to create new rego poilicy")
		return bson.NilObjectID, fmt.Errorf("failed to insert rego")
	}

	insertedID, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		log.Error().Str("function", "Create").Str("Rego ID", rego.ID.Hex()).Msg("failed to convert inserted ID to ObjectID")
		return bson.NilObjectID, fmt.Errorf("failed to convert inserted ID to ObjectID %s", rego.ID.Hex())
	}

	log.Info().Str("function", "Create").Str("jobID", insertedID.Hex()).Msg("rego policy created successfully")
	return insertedID, nil
}

func (r *regoRepository) FindByResourceType(resourceType string) (*RegoPolicy, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rego RegoPolicy

	err := r.collection.FindOne(ctx, bson.M{"resource_type": resourceType}).Decode(&rego)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warn().Str("function", "FindByResourceType").Str("Resoruce Type:", resourceType).Msg("No rego policy found for resource type")
			return nil, fmt.Errorf("no rego policy found for resource type %s", resourceType)
		}
		log.Error().Err(err).Str("function", "FindByResourceType").Str("ResourceType", resourceType).Msg("Failed to execute find")
		return nil, fmt.Errorf("failed to execute find for rego policy by resource type %s", resourceType)
	}

	return &rego, nil

}
