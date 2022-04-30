package handler

import (
	"fmt"
	"github.com/akrennmair/slice"
	"github.com/ccundiff/spend-tracker/transaction-importer/client"
	spendsummaries "github.com/ccundiff/spend-tracker/transaction-importer/spend-summaries"
	"github.com/ccundiff/spend-tracker/transaction-importer/timeutil"
	"github.com/ccundiff/spend-tracker/transaction-importer/transactions"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	f "github.com/fauna/faunadb-go/v4/faunadb"
	"strings"
	"time"
)

type Transaction struct {
	Merchant string
	Amount   float32
}

type DailySpendSummary struct {
	Date       string  `fauna:"date"`
	Month      int     `fauna:"month"`
	SpendGoal  int     `fauna:"spendGoal"`
	UserRef    f.RefV  `fauna:"userRef"`
	TotalSpend float32 `fauna:"totalSpend"`
	SpendDiff  float32 `fauna:"spendDiff"`
}

type TransactionImportHandler struct {
	twilioClient        *client.TwilioClient
	usersService        *users.UserService
	transactionsService *transactions.TransactionsService
	spendSummaryService *spendsummaries.SpendSummariesService
}

func NewTranscationImportHandler(twilioClient *client.TwilioClient,
	usersService *users.UserService, transactionsService *transactions.TransactionsService, spendSummaryService *spendsummaries.SpendSummariesService) *TransactionImportHandler {
	return &TransactionImportHandler{
		twilioClient:        twilioClient,
		usersService:        usersService,
		transactionsService: transactionsService,
		spendSummaryService: spendSummaryService,
	}
}

// todo: need to make importDate a time so I can use it for the month later on
func (t *TransactionImportHandler) Handle() (string, error) {

	user, err := t.usersService.GetUser()
	if err != nil {
		panic(fmt.Sprintf("unable to get user, err=[%v]", err))
	}

	const iso8601TimeFormat = "2006-01-02"
	//startDate := time.Now().Add(-24 * time.Hour).Format(iso8601TimeFormat)
	// TODO: this should be set on the user object
	loc, _ := time.LoadLocation("America/New_York")
	currentTime := time.Now().In(loc)
	println(currentTime.Format(iso8601TimeFormat))

	err = t.transactionsService.ImportTransactions(user, timeutil.EastCoastCurrentDateAsString())
	if err != nil {
		return "error importing txns", err
	}

	dailySpendSummary, txns, err := t.spendSummaryService.CreateDailySpendSummary(user, timeutil.EastCoastCurrentDateAsString())
	if err != nil {
		return "err creating daily spend summary", err
	}

	txnsString := slice.Map(txns, func(value transactions.Transaction) string {
		return fmt.Sprintf("%v : $%v", value.Merchant, value.Amount)
	})

	dailySpendMessage := fmt.Sprintf(
		"Total Day Spend: $%v \n\n"+
			"Day Spend Diff: $%v \n\n"+
			"Transactions: \n"+
			"%v \n\n"+
			"To Date Monthly Surplus: $%v",
		dailySpendSummary.TotalSpend, dailySpendSummary.SpendDiff, strings.Join(txnsString, "\n"), *dailySpendSummary.ToDateMonthSpendDiff,
	)
	println(dailySpendMessage)

	if err = t.twilioClient.SendText("+12692088780", dailySpendMessage); err != nil {
		fmt.Printf("Error sending text, %v", err)
		return "failed to send text", err
	}

	//return fmt.Sprintf("created txns, %v", createTxns), err
	return fmt.Sprintf("created txns"), err
}
