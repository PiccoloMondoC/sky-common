package secrets

import (
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

type SecretFetcher interface {
	GetSecret(key string) (string, error)
}

type EnvVarSecretFetcher struct{}

func (f *EnvVarSecretFetcher) GetSecret(key string) (string, error) {
	// Fetch secret from environment variable
	return os.Getenv(key), nil
}

type GcpSecretManagerFetcher struct {
	client    *secretmanager.Client
	projectID string
}

func NewGcpSecretManagerFetcher(projectID string, credentialsFile string) (*GcpSecretManagerFetcher, error) {
	ctx := context.Background()

	client, err := secretmanager.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to setup secret manager client: %v", err)
	}

	return &GcpSecretManagerFetcher{
		client:    client,
		projectID: projectID,
	}, nil
}

func (f *GcpSecretManagerFetcher) GetSecret(secretID string) (string, error) {
	ctx := context.Background()

	// Build the request.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", f.projectID, secretID),
	}

	// Call the API.
	result, err := f.client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	// Return the secret payload as a string.
	return string(result.Payload.Data), nil
}

func GetFetcher() SecretFetcher {
	// Decide on the method based on a configuration setting or an environment variable.
	// For example, we can use an environment variable "USE_SECRET_MANAGER"
	useSecretManager := os.Getenv("USE_SECRET_MANAGER")
	if useSecretManager == "true" {
		return &GcpSecretManagerFetcher{}
	}
	return &EnvVarSecretFetcher{}
}
