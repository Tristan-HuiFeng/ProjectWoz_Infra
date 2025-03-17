package gcpcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	awscloud "github.com/Tristan-HuiFeng/ProjectWoz_Infra/internal/cloud/aws"
	sigv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go/logging"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// Header represents an HTTP header in the format Google expects
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetCallerIdentityToken represents the signed request for AWS STS GetCallerIdentit
type GetCallerIdentityToken struct {
	URL     string   `json:"url"`
	Method  string   `json:"method"`
	Headers []Header `json:"headers"`
}

// GcpStsTokenExchangeRequest represents the token request to Google STS
type GcpStsTokenExchangeRequest struct {
	Audience            string `json:"audience"`
	GrantType           string `json:"grantType"`
	RequestedTokenType  string `json:"requestedTokenType"`
	Scope               string `json:"scope"`
	SubjectTokenType    string `json:"subjectTokenType"`
	SubjectToken        string `json:"subjectToken"`
}

// GcpStsTokenExchangeResponse represents the token response from Google STS
type GcpStsTokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	IssuedTokenType string `json:"issued_token_type"`
	TokenType string `json:"token_type"`
	ExpiresIn int `json:"expires_in"`
}


func getFederatedTokenFromAws(ctx context.Context, gcpRegisteredIdProvider string) (*GcpStsTokenExchangeResponse, error) {
	// Prepare a GetCallerIdentity request
	url := "https://sts.amazonaws.com/?Action=GetCallerIdentity&Version=2011-06-15"
	getCallerIdentityRequest, err := http.NewRequest("POST", url, nil)
	
	if err != nil {
		fmt.Printf("unable to create request: %v\n", err)
	}

	// Add required headers
	getCallerIdentityRequest.Header.Set("Host", "sts.amazonaws.com")
	getCallerIdentityRequest.Header.Set("x-goog-cloud-target-resource", gcpRegisteredIdProvider)

	// Set session credentials
	cfg, err := awscloud.GetRoleConfig()
	if err != nil {
		fmt.Printf("unable to load AWS SDK config, %v\n", err)
	}

	credsValue, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		fmt.Printf("unable to retrieve credentials: %v\n", err)
	}

	// Create AWS signature V4 signer
	signer := sigv4.NewSigner(func(o *sigv4.SignerOptions) {
		o.Logger = logging.Nop{}
		o.LogSigning = false
	})

	// Sign the request
	// Hex encoded SHA-256 hash of an empty string to be used as as the payloadHash value in signer.SignHTTP
	emptyStringHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	err = signer.SignHTTP(ctx, credsValue, getCallerIdentityRequest, emptyStringHash, "sts", cfg.Region, time.Now())
	if err != nil {
		log.Error().Err(err).Msgf("unable to sign request: %v", err)
	}

	// Prepare GetCallerIdentityToken for Google
	callerIdentityToken := GetCallerIdentityToken{
		URL:    getCallerIdentityRequest.URL.String(),
		Method: getCallerIdentityRequest.Method,
		Headers: []Header{},
	}

	// Add headers to the format Google expects
	for key, values := range getCallerIdentityRequest.Header {
		for _, value := range values {
			callerIdentityToken.Headers = append(callerIdentityToken.Headers, Header{
				Key:   key,
				Value: value,
			})
		}
	}

	callerIdentityTokenJson, err := json.Marshal(callerIdentityToken)
	if err != nil {
		log.Error().Err(err).Msgf("unable to marshal caller identity token: %v", err)
	}

	// Create federated token request body
	federatedTokenRequestBody := GcpStsTokenExchangeRequest{
		Audience:           gcpRegisteredIdProvider,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		Scope:              "https://www.googleapis.com/auth/cloud-platform",
		SubjectTokenType:   "urn:ietf:params:aws:token-type:aws4_request",
		SubjectToken:       string(callerIdentityTokenJson),
	}

	federatedTokenRequestBodyJson, err := json.Marshal(federatedTokenRequestBody)
	if err != nil {
		log.Error().Err(err).Msgf("unable to marshal federated token request body: %v\n", err)
	}

	// Send federated token request to Google STS
	federatedTokenRequest, err := http.NewRequest(
		"POST",
		"https://sts.googleapis.com/v1/token",
		bytes.NewBuffer(federatedTokenRequestBodyJson),
	)

	if err != nil {
		log.Error().Err(err).Msgf("unable to create federated token request: %v\n", err)
	}

	federatedTokenRequest.Header.Set("Content-Type", "application/json")
	federatedTokenResp, err := http.DefaultClient.Do(federatedTokenRequest)
	if err != nil {
		log.Error().Err(err).Msgf("unable to get federated token: %v", err)
	}
	defer federatedTokenResp.Body.Close()

	if federatedTokenResp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(federatedTokenResp.Body).Decode(&errorResponse); err != nil {
			log.Error().Err(err).Msgf("unable to decode error response: %v", err)
		}
		log.Error().Err(err).Msgf("error getting federated token: %v", errorResponse)
	}

	var federatedToken GcpStsTokenExchangeResponse
	if err := json.NewDecoder(federatedTokenResp.Body).Decode(&federatedToken); err != nil {
		log.Error().Err(err).Msgf("unable to decode federated token response: %v", err)
	}

	return &federatedToken, err
}

func createdSignedJwtTokenWithSts(stsToken *GcpStsTokenExchangeResponse, serviceAccount string) (*oauth2.Token, error) {
	// Prepare a GenerateAccessToken request
	url := fmt.Sprintf("https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken", serviceAccount)
	requestBody := map[string]interface{}{
		"scope": []string{
			"https://www.googleapis.com/auth/cloud-platform",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", stsToken.AccessToken))

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response (%d): %s", resp.StatusCode, body)
	}
	
	// Parse the response
	var result struct {
		AccessToken string `json:"accessToken"`
		ExpireTime  string `json:"expireTime"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	
	// Parse expiry time
	expiry, err := time.Parse(time.RFC3339, result.ExpireTime)
	if err != nil {
		expiry = time.Now().Add(time.Hour) // Default to 1 hour if parsing fails
	}
	
	// Create oauth2 token
	token := &oauth2.Token{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		Expiry:      expiry,
	}

	return token, nil
}

func ImpersonateServiceAccount(serviceAccount string) (*oauth2.TokenSource, error) {
	// Hardcoded variables
	ctx := context.Background()

	// Load environment variables
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Error().Err(err).Msg("Error loading .env file")
	}

	gcpRegisteredIdProvider := os.Getenv("GCP_REGISTERED_ID_PROVIDER")

	federatedToken, err := getFederatedTokenFromAws(ctx, gcpRegisteredIdProvider)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting federated token: %v", err)
	}
	saToken, err := createdSignedJwtTokenWithSts(federatedToken, serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to create service account token: %v", err)
	}

	ts := oauth2.StaticTokenSource(saToken)

	return &ts, nil
}