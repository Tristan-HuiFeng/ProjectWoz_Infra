package opa2

import (
	"context"
	"fmt"
	"time"

	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ScanRepository interface {
	InsertMany(scanResults []interface{}) ([]interface{}, error)
}

type scanRepository struct {
	collection *mongo.Collection
}

func NewScanRepository(db database.Service) ScanRepository {
	return &scanRepository{
		collection: db.GetCollection("scan_result"),
	}
}

func (r *scanRepository) InsertMany(scanResults []interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Insert the documents into the MongoDB collection
	insertResult, err := r.collection.InsertMany(ctx, scanResults)
	if err != nil {
		log.Error().Err(err).Str("function", "InsertMany").Msg("Failed to insert scan results")
		return nil, fmt.Errorf("failed to insert scan results: %w", err)
	}

	log.Info().Str("function", "InsertMany").
		Int("insertedCount", len(insertResult.InsertedIDs)).
		Msg("Scan results inserted successfully")

	// Return inserted IDs
	return insertResult.InsertedIDs, nil
}
