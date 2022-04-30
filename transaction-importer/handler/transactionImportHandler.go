package handler

import (
	"fmt"
	"time"

	"github.com/ccundiff/spend-tracker/transaction-importer/client"
	"github.com/ccundiff/spend-tracker/transaction-importer/timeutil"
	"github.com/ccundiff/spend-tracker/transaction-importer/constants"
	"github.com/ccundiff/spend-tracker/transaction-importer/transactions"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	f "github.com/fauna/faunadb-go/v4/faunadb"
	plaid "github.com/plaid/plaid-go/plaid"
)

type Transaction struct {
	Merchant	string
	Amount 		float32
}

type DailySpendSummary struct {
	Date string `fauna:"date"`
	Month int `fauna:"month"`
	SpendGoal int `fauna:"spendGoal"`
	UserRef f.RefV `fauna:"userRef"`
	TotalSpend float32 `fauna:"totalSpend"`
	SpendDiff float32 `fauna:"spendDiff"`
}

type TransactionImportHandler struct {
	DbClient    *f.FaunaClient
	PlaidClient *plaid.APIClient
	TwilioClient *client.TwilioClient
	UsersService *users.UserService
	transactionsService *transactions.TransactionsService
}

func NewTranscationImportHandler(dbClient *f.FaunaClient, plaidClient *plaid.APIClient, twilioClient *client.TwilioClient,
	usersService *users.UserService, transactionsService *transactions.TransactionsService) *TransactionImportHandler {
	return &TransactionImportHandler{
		DbClient:    dbClient,
		PlaidClient: plaidClient,
		TwilioClient: twilioClient,
		UsersService: usersService,
		transactionsService: transactionsService,
	}
}

// todo: need to make importDate a time so I can use it for the month later on
func (t *TransactionImportHandler) Handle() (string, error) {

	user, err := t.UsersService.GetUser()
	if err != nil {
		panic(fmt.Sprintf("unable to get user, err=[%v]", err))
	}

	const iso8601TimeFormat = "2006-01-02"
	//startDate := time.Now().Add(-24 * time.Hour).Format(iso8601TimeFormat)
	// TODO: this should be set on the user object
	loc, _ := time.LoadLocation("America/New_York")
	currentTime := time.Now().In(loc)
	println(currentTime.Format(iso8601TimeFormat))

	err = t.transactionsService.ImportTransactions(user, timeutil.CurrentDateAsString(constants.EAST_COAST_TIME_LOCATION))
	if err != nil {
		return "error importing txns", err
	}

	_, err = t.transactionsService.TransactionsForDate(timeutil.CurrentDateAsString(constants.EAST_COAST_TIME_LOCATION))
	if err != nil {
		return "err obtaining txns for date", err
	}


	//createStatements := make([]f.Expr, 5)
	//
	//spendTotal := float32(0)
	//// todo: account for categories
	//
	//spendDiff := float32(user.DailySpendGoal) - spendTotal
	//
	//dailySpendSummary := DailySpendSummary{
	//	Date:  currentTime.Format(iso8601TimeFormat),
	//	Month: int(currentTime.Month()),
	//	SpendGoal: user.DailySpendGoal,
	//	UserRef: user.Ref,
	//	TotalSpend: spendTotal,
	//	SpendDiff: spendDiff,
	//}
	//
	//fmt.Printf("%+v\n",dailySpendSummary)

	//createStatements = append(createStatements,
	//	f.Let().Bind(
	//		"dssMatch", f.MatchTerm(f.Index("daily_spend_summaries_by_date"), dailySpendSummary.Date),
	//	).In(
	//		f.If(f.IsNonEmpty(f.Var("dssMatch")),
	//			f.Update(
	//				f.Select("ref", f.Get(f.Var("dssMatch"))), f.Obj{"data": dailySpendSummary},
	//			),
	//			f.Create(f.Collection("DailySpendSummaries"), f.Obj{
	//				"data": dailySpendSummary,
	//			}),
	//		),
	//	))
	//
	//_, err = t.DbClient.Query(
	//	f.Do(createStatements),
	//)
	//if err != nil {
	//	fmt.Printf("error creating txns and spend summary, err=[%v]", err)
	//	panic(fmt.Sprintf("error creating txns and spend summary, err=[%v]", err))
	//}
	//
	//readResp, err := t.DbClient.Query(
	//	f.Select("data",
	//		f.Reduce(f.Lambda(f.Arr{"acc", "val"}, f.Add(f.Var("acc"), f.Var("val"))),
	//			0,
	//			f.Map(f.Paginate(f.MatchTerm(f.Index("daily_spend_summaries_by_month"), dailySpendSummary.Month)),
	//				f.Lambda("ref", f.Select(f.Arr{"data", "spendDiff"}, f.Get(f.Var("ref"))))),
	//		),
	//	),
	//)
	//if err != nil {
	//	fmt.Printf("summing spend, err=[%v]", err)
	//	panic(fmt.Sprintf("summing spend, err=[%v]", err))
	//}
	//
	//var runningSpendDiff []float32
	//if err = readResp.Get(&runningSpendDiff); err != nil {
	//	panic(fmt.Sprintf("error reading spend diff from response, %v", err))
	//}
	//
	//fmt.Printf("totalled spenndiff %v", runningSpendDiff)
	//
	//txnsString := slice.Map(txns, func(value Transaction) string {
	//	return fmt.Sprintf("%v : $%v", value.Merchant, value.Amount)
	//})
	//
	//dailySpendMessage := fmt.Sprintf(
	//	"Total Day Spend: $%v \n\n"+
	//		"Day Spend Diff: $%v \n\n"+
	//		"Transactions: \n"+
	//		"%v \n\n"+
	//		"Month Spend Diff: $%v",
	//	spendTotal, spendDiff, strings.Join(txnsString, "\n"), runningSpendDiff[0],
	//)
	//println(dailySpendMessage)
	//
	//if err = t.TwilioClient.SendText("+12692088780", dailySpendMessage); err != nil {
	//	fmt.Printf("Error sending text, %v", err)
	//	return "failed to send text", err
	//}

	//return fmt.Sprintf("created txns, %v", createTxns), err
	return fmt.Sprintf("created txns"), err
}
