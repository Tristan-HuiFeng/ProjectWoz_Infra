package cloud

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DiscoveryJob struct {
	ID        bson.ObjectID       `bson:"_id,omitempty"` // Unique identifier
	ClientID  string              `bson:"client_id"`     // internal client ID
	AccountID string              `bson:"account_id"`    // GCP project id or AWS account ID
	Status    string              `bson:"status"`        // Job status (e.g., "pending", "in-progress", "completed")
	Resources map[string][]string `bson:"resources"`     // Resources' identifier
	Provider  string              `bson:"provider"`      // Cloud Provider
	CreatedAt int64               `bson:"created_at"`    // Timestamp for job creation
}

// type RetrivalJob struct {
// 	ID             primitive.ObjectID  `bson:"_id,omitempty"` // Unique identifier
// 	Status         string              `bson:"status"`        // Job status (e.g., "pending", "in-progress", "completed")
// 	Resources      map[string][]string `bson:"resources"`
// 	CreatedAt      int64               `bson:"created_at"` // Timestamp for job creation
// 	DiscoveryJobID primitive.ObjectID
// }

// type S3BucketConfig struct {
// 	ID             primitive.ObjectID `bson:"_id,omitempty"`
// 	DiscoveryJobID primitive.ObjectID `bson:"discovery_job_id"`
// 	BucketName     string             `bson:"bucket_name"`
// 	Policy         string             `bson:"policy"`
// 	PublicAccess   bool               `bson:"public_access"`
// 	RetrievedAt    int64              `bson:"retrieved_at"`
// }

type ResourceConfig struct {
	ID             bson.ObjectID          `bson:"_id,omitempty"`
	DiscoveryJobID bson.ObjectID          `bson:"discovery_job_id"` // Link to the discovery job
	ClientID       string                 `bson:"client_id"`        // internal client ID
	AccountID      string                 `bson:"account_id"`       // GCP project id or AWS account ID
	Provider       string                 `bson:"provider"`         // Cloud Provider
	ResourceType   string                 `bson:"resource_type"`    // e.g., "s3", "ec2", "rds", "gcs"
	ResourceID     string                 `bson:"resource_id"`      // e.g., S3 bucket name, EC2 instance ARN
	Config         map[string]interface{} `bson:"config"`           // The actual configuration (could be S3 policy, EC2 security groups, etc.)
}

type BucketConfig struct {
	BucketPolicy string `json:"policy"`
}
