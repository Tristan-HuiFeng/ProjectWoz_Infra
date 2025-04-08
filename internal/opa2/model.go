package opa2

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ScanResult struct {
	ID               bson.ObjectID `bson:"_id,omitempty"`
	DiscoveryJobID   bson.ObjectID `bson:"discovery_job_id"` // Link to the discovery job
	ResourceType     string        `bson:"resource_type"`    // e.g., "s3", "ec2", "rds", "gcs"
	ResourceID       string        `bson:"resource_id"`      // e.g., S3 bucket name, EC2 instance ARN
	Status           string        `bson:"status"`           // status of the Job
	Pass             bool          `bson:"pass"`             // Fixed type from 'boolean' to 'bool'
	Misconfiguration []string      `bson:"misconfiguration"` // misconfiguration details
	ClientID         string        `bson:"client_id"`
	ResourceOwnerID  string        `bson:"resource_owner_id"` // GCP project id or AWS account ID
	Provider         string        `bson:"provider"`          // Cloud Provider
}

type RegoPolicy struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	ResourceType string        `bson:"resource_type"` // e.g., "s3", "ec2", "rds", "gcs"
	Query        string        `bson:"query"`
	Rego         string        `bson:"rego"` // rego policy for a specific resource type
}
