package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/models"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Service interface {
	Health() map[string]string
	GetCollection(name string) *mongo.Collection
	Disconnect() error

	// UserDAL
	CreateUser(context.Context, models.User) (bson.ObjectID, error)
	FetchUserByID(context.Context, bson.ObjectID) (models.User, error)
	FetchUserByName(context.Context, string) (models.User, error)
	FetchUserByEmail(context.Context, string) (models.User, error)
	FetchUsers(context.Context) ([]models.User, error)
	UpdateUser(context.Context, bson.ObjectID, models.User) error
	DeleteUser(context.Context, bson.ObjectID) error
}

type service struct {
	db *mongo.Client
}

var (
	instance Service
	once     sync.Once
)

func New() (Service, error) {

	uri := os.Getenv("MONGO_DB_STRING")
	if uri == "" {
		return nil, fmt.Errorf("database URI not set in environment variables")
	}

	var err error
	once.Do(func() { // Ensures only one instance is created
		client, connErr := mongo.Connect(options.Client().ApplyURI(uri))
		if connErr != nil {
			err = fmt.Errorf("failed to connect to MongoDB: %w", connErr)
			return
		}
		instance = &service{db: client}
		log.Println("Connected to MongoDB!")
	})

	if err != nil {
		return nil, err
	}
	return instance, nil

}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Printf("DB health check failed: %v", err)
		return map[string]string{
			"message": fmt.Sprintf("Database is down: %v", err),
		}
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

func (s *service) GetCollection(name string) *mongo.Collection {
	return s.db.Database("cs464-main").Collection(name)
}

func (s *service) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.db.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("Successfully disconnected from MongoDB")
	return nil
}
