package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ccundiff/spend-tracker/transaction-importer/client"
	"github.com/ccundiff/spend-tracker/transaction-importer/config"
	"github.com/ccundiff/spend-tracker/transaction-importer/handler"
	spendsummaries "github.com/ccundiff/spend-tracker/transaction-importer/spend-summaries"
	"github.com/ccundiff/spend-tracker/transaction-importer/transactions"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	f "github.com/fauna/faunadb-go/v4/faunadb"
	"github.com/plaid/plaid-go/plaid"
)

var transactionImportHandler *handler.TransactionImportHandler

func init() {
	secretsReceiver := config.NewRetriever(context.TODO())
	secretsConfig, err := secretsReceiver.RetrieveSecretsConfig()
	if err != nil {
		panic(fmt.Sprintf("unable to retrieve secrets config, err=[%v]", err))
	}

	faunaClient := f.NewFaunaClient(
		secretsConfig.FaunaDBKey,
		f.Endpoint("https://db.us.fauna.com"),
	)

	plaidConfig := plaid.NewConfiguration()
	plaidConfig.AddDefaultHeader("PLAID-CLIENT-ID", secretsConfig.PlaidClientId)
	plaidConfig.AddDefaultHeader("PLAID-SECRET", secretsConfig.PlaidKey)
	plaidConfig.UseEnvironment(plaid.Development)
	plaidClient := plaid.NewAPIClient(plaidConfig)
	twilioClient := client.NewTwilioClient(secretsConfig.TwilioAccountSid, secretsConfig.TwilioAuthToken)

	transactionsService := transactions.NewTransactionsService(
		faunaClient,
		plaidClient,
	)
	usersService := users.NewUsersService(faunaClient)
	spendSummariesService := spendsummaries.NewSpendSummariesService(faunaClient, transactionsService)
	transactionImportHandler = handler.NewTranscationImportHandler(twilioClient, usersService, transactionsService, spendSummariesService)
}

func HandleRequest(ctx context.Context) (string, error) {
	return transactionImportHandler.Handle()
}

func main() {
	lambda.Start(HandleRequest)
}
