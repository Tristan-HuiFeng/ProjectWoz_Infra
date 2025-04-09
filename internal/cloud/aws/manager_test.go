package awscloud_test

import (
	"testing"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"

	cloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MockDiscoveryRepository struct {
	mock.Mock
}

func (m *MockDiscoveryRepository) Create(job *cloud.DiscoveryJob, clientID string, resourceOwnerID string) (bson.ObjectID, error) {
	args := m.Called(job)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *MockDiscoveryRepository) FindByID(id bson.ObjectID) (*cloud.DiscoveryJob, error) {
	args := m.Called(id)
	return args.Get(0).(*cloud.DiscoveryJob), args.Error(1)
}

func (m *MockDiscoveryRepository) UpdateResources(id bson.ObjectID, resources map[string][]string) error {
	args := m.Called(id, resources)
	return args.Error(0)
}

func (m *MockDiscoveryRepository) UpdateJob(id bson.ObjectID, resourceName string, resourceData []string) error {
	args := m.Called(id, resourceName, resourceData)
	return args.Error(0)
}

func (m *MockDiscoveryRepository) UpdateStatus(id bson.ObjectID, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) Create(job *awscloud.ConfigRepository) error {
	args := m.Called(job)
	return args.Error(1)
}

func (m *MockConfigRepository) InsertMany(resourceConfigs []interface{}) ([]interface{}, error) {
	args := m.Called(resourceConfigs)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockConfigRepository) FindByDiscoveryID(discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error) {
	args := m.Called(discoveryJobID)
	return args.Get(0).([]cloud.ResourceConfig), args.Error(1)
}

func (m *MockConfigRepository) FindByTypeAndJobID(resourceType string, discoveryJobID bson.ObjectID) ([]cloud.ResourceConfig, error) {
	args := m.Called(resourceType, discoveryJobID)
	return args.Get(0).([]cloud.ResourceConfig), args.Error(1)
}

type MockResource struct {
	mock.Mock
}

func (m *MockResource) Discover(cfg aws.Config) ([]string, error) {
	args := m.Called(cfg)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockResource) RetrieveConfig(cfg aws.Config, resourceIDs []string) (map[string]map[string]interface{}, error) {
	args := m.Called(cfg, resourceIDs)
	return args.Get(0).(map[string]map[string]interface{}), args.Error(1)
}

func (m *MockResource) Name() string {
	args := m.Called()
	return args.String(0)
}

func TestRunDiscovery(t *testing.T) {
	// Setup
	mockDiscoveryRepo := new(MockDiscoveryRepository)
	mockResource := new(MockResource)
	cfg := aws.Config{}

	// Mock the methods
	jobID := bson.NewObjectID()

	mockDiscoveryRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(jobID, nil)
	mockDiscoveryRepo.On("UpdateJob", jobID, "s3", mock.Anything).Return(nil)
	mockDiscoveryRepo.On("UpdateStatus", jobID, "completed").Return(nil)

	mockResource.On("Discover", cfg).Return([]string{"resource1", "resource2"}, nil)
	mockResource.On("Name").Return("s3")

	// Call RunDiscovery
	returnedJobID, err := awscloud.RunDiscovery(cfg, mockDiscoveryRepo, "123", []awscloud.ResourceDiscovery{mockResource}, "123")

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, jobID, returnedJobID)

	mockDiscoveryRepo.AssertExpectations(t)
	mockResource.AssertExpectations(t)
}

// Test Retrival
func TestRetrival(t *testing.T) {
	// Setup
	mockDiscoveryRepo := new(MockDiscoveryRepository)
	mockConfigRepo := new(MockConfigRepository)
	mockResource := new(MockResource)
	cfg := aws.Config{}
	jobID := bson.NewObjectID()

	// Mock the methods
	discoveryJob := &cloud.DiscoveryJob{
		ID:        jobID,
		Resources: map[string][]string{"s3": {"resource1", "resource2"}},
	}

	resourceConfigs := map[string]map[string]interface{}{
		"resource1": {"policy": "read-only"},
		"resource2": {"policy": "admin"},
	}

	mockDiscoveryRepo.On("FindByID", jobID).Return(discoveryJob, nil)
	mockResource.On("RetrieveConfig", cfg, []string{"resource1", "resource2"}).Return(resourceConfigs, nil)
	mockConfigRepo.On("InsertMany", mock.Anything).Return([]interface{}{"inserted1", "inserted2"}, nil)

	mockResource.On("Name").Return("s3")

	// Call Retrival
	err := awscloud.RunRetrieval(cfg, mockDiscoveryRepo, mockConfigRepo, jobID, []awscloud.ResourceDiscovery{mockResource}, "123", "123")

	// Assertions
	assert.NoError(t, err)

	mockDiscoveryRepo.AssertExpectations(t)
	mockResource.AssertExpectations(t)
	mockConfigRepo.AssertExpectations(t)
}
