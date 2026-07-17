package main

import (
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	sm "github.com/IBM/secrets-manager-go-sdk/v2/secretsmanagerv2"
)

// IAMAPIKeyFromSecretsManager fetches an iam_credentials secret from IBM
// Secrets Manager and returns the vended IAM API key string.
//
// The secret identified by cfg.SecretID must be of type iam_credentials.
func IAMAPIKeyFromSecretsManager(cfg SecretsManagerConfig) (string, error) {
	client, err := sm.NewSecretsManagerV2(&sm.SecretsManagerV2Options{
		URL: cfg.InstanceURL,
		Authenticator: &core.IamAuthenticator{
			ApiKey: cfg.APIKey,
		},
	})
	if err != nil {
		return "", fmt.Errorf("creating Secrets Manager client: %w", err)
	}

	result, _, err := client.GetSecret(client.NewGetSecretOptions(cfg.SecretID))
	if err != nil {
		return "", fmt.Errorf("GetSecret %s: %w", cfg.SecretID, err)
	}

	iamSecret, ok := result.(*sm.IAMCredentialsSecret)
	if !ok {
		return "", fmt.Errorf("secret %s is not an iam_credentials secret", cfg.SecretID)
	}
	if iamSecret.ApiKey == nil || *iamSecret.ApiKey == "" {
		return "", fmt.Errorf("secret %s has no api_key value (may not have been retrieved yet)", cfg.SecretID)
	}

	return *iamSecret.ApiKey, nil
}
