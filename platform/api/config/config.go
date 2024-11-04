package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type Config struct {
	ServerPort         string `json:"PORT"`
	DBConnString string `json:"CONN_STRING"`
	Auth0Secret string `json:"AUTH0_SECRET"`
	Auth0Domain string `json:"AUTH0_DOMAIN"`
	Auth0BaseURL string `json:"AUTH0_BASE_URL"`
	Auth0IssuerBaseURL string `json:"AUTH0_ISSUER_BASE_URL"`
	Auth0ClientID string `json:"AUTH0_CLIENT_ID"`
	Auth0ClientSecret string `json:"AUTH0_CLIENT_SECRET"`
	Auth0RoleID string `json:"AUTH0_ROLE_ID"`
	Auth0Audience string `json:"AUTH0_AUDIENCE"`
	Auth0HookSecret string `json:"AUTH_HOOK_SECRET"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "local"
	}

	region := os.Getenv("AWS_REGION")

	secretName := fmt.Sprintf("%s-traba-backend-config", env)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValueWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %v", err)
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		return nil, fmt.Errorf("secret value is not a string")
	}

	var config Config
	err = json.Unmarshal([]byte(secretString), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret value: %v", err)
	}

	return &config, nil
}


