package handler

import (
	"fmt"
	"github.com/akrennmair/slice"
	"github.com/ccundiff/spend-tracker/transaction-importer/client"
	spendsummaries "github.com/ccundiff/spend-tracker/transaction-importer/spend-summaries"
	"github.com/ccundiff/spend-tracker/transaction-importer/timeutil"
	"github.com/ccundiff/spend-tracker/transaction-importer/transactions"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	"strings"
)

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

	// TODO: timezone should be set on the user object
	// could also just run it for previous day and not really worry about timezones
	// could have user set their desired notificaiton time
	err = t.transactionsService.ImportTransactions(user, timeutil.EastCoastYesterdaysDateAsString())
	if err != nil {
		return "error importing txns", err
	}

	dailySpendSummary, txns, err := t.spendSummaryService.CreateDailySpendSummary(user, timeutil.EastCoastYesterdaysDateAsString())
	if err != nil {
		return "err creating daily spend summary", err
	}

	txnsString := slice.Map(txns, func(value transactions.Transaction) string {
		return fmt.Sprintf("%v : $%v", value.Merchant, value.Amount)
	})

	toDateMonthSurplus := user.MonthlyIncome - *dailySpendSummary.ToDateMonthSpend

	dailySpendMessage := fmt.Sprintf(
		"Total Day Spend on %v: $%v \n\n"+
			"Day Spend Diff: $%v \n\n"+
			"Transactions: \n"+
			"%v \n\n"+
			"To Date Monthly Surplus: $%v",
		timeutil.EastCoastYesterdaysDateAsString(), dailySpendSummary.TotalSpend, dailySpendSummary.SpendDiff, strings.Join(txnsString, "\n"), toDateMonthSurplus,
	)
	println(dailySpendMessage)

	// TODO: phone number stored with user
	if err = t.twilioClient.SendText("+12692088780", dailySpendMessage); err != nil {
		fmt.Printf("Error sending text, %v", err)
		return "failed to send text", err
	}

	return fmt.Sprintf("created txns"), err
}
