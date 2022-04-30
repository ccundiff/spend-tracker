package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	//"github.com/aws/aws-sdk-go-v2/aws/session"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const serviceSecretsName = "TransactionImporter"

type Configuration struct {
	FaunaDBKey string `json:"fauna-db-key"`
	PlaidKey string `json:"plaid-key"`
	PlaidClientId string `json:"plaid-client-id"`
	TwilioAccountSid string `json:"twilio-account-sid"`
	TwilioAuthToken string `json:"twilio-auth-token"`
}

type Retriever interface {
	RetrieveSecretsConfig() (Configuration, error)
}

func NewRetriever(ctx context.Context) Retriever {
	cfg, _ := config.LoadDefaultConfig(ctx)
	svc := secretsmanager.NewFromConfig(cfg)

	return &awsSecretsRetriever{secretsClient: svc}
}

type awsSecretsRetriever struct {
	ctx context.Context
	secretsClient *secretsmanager.Client
}

func (awsSecretsRetriever *awsSecretsRetriever) RetrieveSecretsConfig() (Configuration, error) {
	secrets, err := awsSecretsRetriever.secretsClient.GetSecretValue(context.TODO(),
		&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(serviceSecretsName),
	})
	if err != nil {
		return Configuration{}, fmt.Errorf("unexpected error attempting to retrieve secrets from aws, "+
			" err=[%v]", err)
	}

	if secrets.SecretString == nil {
		return Configuration{}, fmt.Errorf("expected secret with name=[%v] to have a defined SecretString", serviceSecretsName)
	}

	var secretsConfig Configuration
	err = json.Unmarshal([]byte(*secrets.SecretString), &secretsConfig)
	if err != nil {
		return Configuration{}, fmt.Errorf("unexpected error attempting to unmarshal secrets, "+
			"err=[%w]", err)
	}

	return secretsConfig, nil
}
