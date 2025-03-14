package opa2

import (
	"context"
	"fmt"

	"github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud"
	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func RunScan(configRepo awscloud.ConfigRepository, scanRepo ScanRepository, regoRepo RegoRepository, discoveryID bson.ObjectID, resources []awscloud.ResourceDiscovery) error {
	log.Info().Str("Discovery ID", discoveryID.Hex()).Msg("Starting misconfig scan")

	for _, resource := range resources {
		log.Info().Str("Discovery ID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Running misconfig scan")

		configs, err := configRepo.FindByTypeAndJobID(resource.Name(), discoveryID)
		if err != nil {
			log.Warn().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Failed to find config")
			continue
		}

		// policyPaths, query, err := setupScan(resource.Name())
		// if err != nil {
		// 	log.Warn().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Failed to get policy")
		// 	continue
		// }

		regoPolicy, err := regoRepo.FindByResourceType("s3")
		if err != nil {
			log.Warn().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Failed to get rego policy")
			continue
		}

		log.Info().Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Str("Rego Query", regoPolicy.Query).Str("Rego Policy", regoPolicy.Rego).Msg("rego debug")

		var scanResults []interface{}

		for _, config := range configs {

			log.Info().Str("function", "EvaluateConfig").Str("resource name", config.ResourceID).Msg("Running evaluation for specific resource")

			misconfigResult, err := EvaluateConfig(regoPolicy, config.Config)
			status := "completed"
			if err != nil {
				log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Failed to get scan result")
				status = "error"
			}

			scanResult := ScanResult{
				DiscoveryJobID:   discoveryID,
				ResourceType:     resource.Name(),
				ResourceID:       config.ResourceID,
				Status:           status,
				Pass:             len(misconfigResult) == 0,
				Misconfiguration: misconfigResult,
			}

			scanResults = append(scanResults, scanResult)

		}

		result, err := scanRepo.InsertMany(scanResults)

		if err != nil {
			log.Error().Err(err).Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Msg("Failed to insert scan result")
			return fmt.Errorf("RunScan: %w", err)
		}
		log.Info().Str("discoveryID", discoveryID.Hex()).Str("resource", resource.Name()).Int("inserted", len(result)).Msg("Scan result inserted successfully")

	}

	return nil
}

func setupScan(name string) ([]string, string, error) {
	var policyPaths []string
	var query string

	switch name {
	case "s3":
		policyPaths = []string{"../../internal/opa2/policies/aws/s3.rego"}
		query = "data.s3.deny"
	default:
		return nil, "", fmt.Errorf("unsupport resource: %s", name)
	}

	return policyPaths, query, nil
}

func EvaluateConfig(regoPolicy *RegoPolicy, config map[string]interface{}) ([]string, error) {
	// func EvaluateConfig(policyPaths []string, query string, config map[string]interface{}) ([]string, error) {
	ctx := context.TODO()

	rq, err := rego.New(
		rego.Query(regoPolicy.Query),
		rego.Module("s3.rego", regoPolicy.Rego),
	).PrepareForEval(ctx)

	if err != nil {
		// Handle error.
		log.Error().Err(err).Str("function", "EvaluateConfig").Msg("Failed to prepare OPA query")
		return []string{}, err
	}

	results, err := rq.Eval(context.Background(), rego.EvalInput(config))
	if err != nil {
		log.Error().Err(err).Str("function", "EvaluateConfig").Msg("Failed to evaluate OPA query")
	}

	log.Debug().Err(err).Str("function", "EvaluateConfig").Msg("Debug")
	for _, r := range results {
		for _, e := range r.Expressions {
			log.Info().
				Str("expression_text", e.Text).
				Interface("expression_value", e.Value).
				Msg("Expression result")
		}
	}

	var misconfiguration []string

	if exp, ok := results[0].Expressions[0].Value.(map[string]interface{}); ok {
		for k, v := range exp {
			misconfiguration = append(misconfiguration, k)
			log.Info().
				Str("expression_text", k).
				Interface("expression_value", v).
				Msg("Expression result")
		}
	}

	return misconfiguration, nil

}

func EvaluateS3BucketPolicy(config cloud.ResourceConfig, regoPolicy *RegoPolicy) (string, error) {

	ctx := context.TODO()

	policyPaths := []string{"../../internal/opa2/policies/aws/s3.rego"}
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return "", err
	// }

	// fmt.Println("Current working directory:", cwd)

	// var inputData map[string]interface{}
	// err := json.Unmarshal([]byte(bucketPolicy), &inputData)
	// if err != nil {
	// 	log.Info().Msgf("Error unmarshalling input: %s", err)
	// 	return "", err
	// }

	// if bucketPolicyJSON, ok := config.Config["bucket_policy"].(string); ok {
	// 	var bucketPolicy map[string]interface{}
	// 	err := json.Unmarshal([]byte(bucketPolicyJSON), &bucketPolicy)
	// 	if err != nil {
	// 		log.Error().Err(err).Msg("Error unmarshalling bucket_policy JSON")
	// 		return "", err
	// 	}

	// 	// Replace the bucket_policy field with the unmarshalled map
	// 	config.Config["bucket_policy"] = bucketPolicy
	// } else {
	// 	// If bucket_policy is already a map, no need to unmarshal
	// 	log.Info().Msg("bucket_policy is already a map, skipping unmarshal.")
	// }

	query, err := rego.New(
		rego.Query("data.s3.deny"),
		rego.Load(policyPaths, nil),
		//rego.Module("s3.rego", regoPolicy.Rego),
	).PrepareForEval(ctx)

	if err != nil {
		// Handle error.
		log.Error().Err(err).Str("function", "EvaluateS3BucketPolicy").Msg("Failed to prepare OPA query")
		return "", err
	}
	fmt.Println("HALOOOOOOOOOO")

	// fmt.Println("\n")
	// for k, v := range config.Config {
	// 	fmt.Printf("%s: %s", k, v)
	// }
	// fmt.Println("\n")

	results, err := query.Eval(context.Background(), rego.EvalInput(config.Config))
	if err != nil {
		log.Error().Err(err).Str("function", "EvaluateS3BucketPolicy").Msg("Failed to evaluate OPA query")
	}

	for _, r := range results {
		for _, e := range r.Expressions {
			log.Info().
				Str("expression_text", e.Text).
				Interface("expression_value", e.Value).
				Msg("Expression result")
		}
		break
	}

	if valueMap, ok := results[0].Expressions[0].Value.(map[string]interface{}); ok {
		for key, value := range valueMap {
			// Now you can range over the map
			log.Info().
				Str("key", key).
				Interface("value", value).
				Msg("Expression result")
		}
	}

	return "", nil

}
