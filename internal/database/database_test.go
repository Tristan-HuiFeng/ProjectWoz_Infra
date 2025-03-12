package database_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	database "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/database"
	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/models"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func connectToMongo() (*mongo.Client, error) {
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Printf("Error loading .env file %v\n", err)
	}

	uri := os.Getenv("BLUEPRINT_TEST_DB_DATABASE")
	if uri == "" {
		return nil, fmt.Errorf("database URI not set in environment variables")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	return client, nil
}

func TestMain(m *testing.M) {
	client, err := connectToMongo()
	if err != nil {
		log.Fatalf("could not start mongodb container: %v", err)
	}

	code := m.Run()

	if client != nil && client.Disconnect(context.Background()) != nil {
		log.Fatalf("could not teardown mongodb container: %v", err)
	}

	os.Exit(code)
}

func TestNew(t *testing.T) {
	_, err := database.New()
	if err != nil {
		t.Fatal("database.New() returned nil")
	}
}

func TestHealth(t *testing.T) {
	srv, err := database.New()
	if err != nil {
		t.Fatal("database.New() returned nil")
	}

	stats := srv.Health()

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestUserDal(t *testing.T) {
	srv, err := database.New()
	if err != nil {
		t.Fatal("database.New() returned nil")
	}

	collection := srv.GetCollection("users")
	if collection == nil {
		t.Fatal("GetCollection() returned nil")
	}

	sampleUser := &models.User{
		Name:           "Test User",
		Email:          "test_user@testuser.com",
		OrganisationID: bson.NewObjectID(),
	}

	// Test CreateUser
	newUserId, err := srv.CreateUser(context.Background(), *sampleUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("newUserId: %v\n", newUserId)

	// Test FetchUserByID
	fetchedUser, err := srv.FetchUserByID(context.Background(), newUserId)
	if err != nil {
		t.Fatalf("Failed to fetch user by ID: %v", err)
	}
	fmt.Printf("fetchedUser: %v\n", fetchedUser)

	// Test FetchUserByName
	fetchedUser, err = srv.FetchUserByName(context.Background(), "Test User")
	if err != nil {
		t.Fatalf("Failed to fetch user by name: %v", err)
	}
	fmt.Printf("fetchedUser: %v\n", fetchedUser)

	// Test FetchUserByEmail
	fetchedUser, err = srv.FetchUserByEmail(context.Background(), "test_user@testuser.com")
	if err != nil {
		t.Fatalf("Failed to fetch user by email: %v", err)
	}
	fmt.Printf("fetchedUser: %v\n", fetchedUser)

	// Test UpdateUser
	newUser := &models.User{
		Name:           "Updated User",
		Email:          "updated_user@testuser.com",
		OrganisationID: fetchedUser.OrganisationID,
	}

	err = srv.UpdateUser(context.Background(), newUserId, *newUser)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	fmt.Println("User updated successfully")
	fmt.Printf("newUser: %v\n", newUser)

	// Test DeleteUser
	err = srv.DeleteUser(context.Background(), newUserId)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Println("User deleted successfully")
}
