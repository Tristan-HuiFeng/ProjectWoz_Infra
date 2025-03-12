package database

// Data Access Layer for User Objects

import (
	"context"
	"fmt"

	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrUserNotFound  = fmt.Errorf("user not found")
	ErrInvalidUserID = fmt.Errorf("invalid user ID")
)

func (s *service) CreateUser(ctx context.Context, user models.User) (bson.ObjectID, error) {
	collection := instance.GetCollection("users")

	// TODO: Validate user object and OrgnisationID

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("failed to create user: %w", err)
	}

	return result.InsertedID.(bson.ObjectID), nil
}

func (s *service) FetchUserByID(ctx context.Context, id bson.ObjectID) (models.User, error) {
	var user models.User
	collection := instance.GetCollection("users")
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

func (s *service) FetchUserByName(ctx context.Context, name string) (models.User, error) {
	var user models.User
	collection := instance.GetCollection("users")
	err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

func (s *service) FetchUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	collection := instance.GetCollection("users")
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("failed to fetch user: %w", err)
	}

	return user, nil
}

func (s *service) FetchUsers(ctx context.Context) ([]models.User, error) {
	collection := instance.GetCollection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (s *service) UpdateUser(ctx context.Context, id bson.ObjectID, user models.User) error {
	collection := instance.GetCollection("users")
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": user},
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *service) DeleteUser(ctx context.Context, id bson.ObjectID) error {
	collection := instance.GetCollection("users")
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}
