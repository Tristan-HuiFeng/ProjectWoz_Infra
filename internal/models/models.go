package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Entity struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
}

type User struct {
	Name           string        `bson:"name,omitempty"`
	Email          string        `bson:"email,omitempty"`
	OrganisationID bson.ObjectID `bson:"organisation_id,omitempty"`
}

type UserEntity struct {
	Entity
	User
}
