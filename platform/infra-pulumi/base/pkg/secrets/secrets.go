package secrets

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/secretsmanager"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AwsSecrets struct {
	Environment string
	Service     string
}

// NewBackendConfig creates ONLY the secret container:
// fs0ciety-<env>-config
//
// No SecretVersion
// No key/value pairs
// Nothing stored yet
func NewAwsSecrets(ctx *pulumi.Context, name string, config AwsSecrets) (*secretsmanager.Secret, error) {

	secretName := fmt.Sprintf("%s-%s-%s-config", name, config.Environment, config.Service)

	secret, err := secretsmanager.NewSecret(ctx, secretName, &secretsmanager.SecretArgs{
		Name: pulumi.String(secretName),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String(config.Environment),
			"Service":     pulumi.String(config.Service),
		},
	})
	if err != nil {
		return nil, err
	}

	return secret, nil
}
