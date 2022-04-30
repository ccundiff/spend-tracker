package spendsummaries

import (
	"fmt"
	"github.com/ccundiff/spend-tracker/transaction-importer/timeutil"
	"github.com/ccundiff/spend-tracker/transaction-importer/transactions"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	f "github.com/fauna/faunadb-go/v4/faunadb"
	"github.com/pkg/errors"
)

type SpendSummariesService struct {
	dbClient            *f.FaunaClient
	transactionsService *transactions.TransactionsService
}

func NewSpendSummariesService(dbClient *f.FaunaClient, transactionsService *transactions.TransactionsService) *SpendSummariesService {
	return &SpendSummariesService{
		dbClient:            dbClient,
		transactionsService: transactionsService,
	}
}

func (s *SpendSummariesService) CreateDailySpendSummary(user users.User, date string) (DailySpendSummary, []transactions.Transaction, error) {
	txns, err := s.transactionsService.TransactionsForDate(date)
	if err != nil {
		return DailySpendSummary{}, nil, errors.Wrapf(err, "Failed to pull txn when creating daily spend summaries, err=%v", err)
	}
	spendTotal := float32(0)
	for _, txn := range txns {
		fmt.Printf("Txn %v=\n", txn)
		spendTotal += txn.Amount
	}

	runningSpendDiff, err := s.getToDateMonthlySpend(date)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}

	month, err := timeutil.GetMonthFromDateString(date)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}

	dailySpendSummary := DailySpendSummary{
		Date:                 date,
		Month:                month,
		SpendGoal:            user.DailySpendGoal,
		UserId:               user.Id,
		TotalSpend:           spendTotal,
		SpendDiff:            float32(user.DailySpendGoal) - spendTotal,
		ToDateMonthSpendDiff: &runningSpendDiff,
	}

	// TODO: need to account for user id here as well....
	_, err = s.dbClient.Query(
		f.Let().Bind(
			"dssMatch", f.MatchTerm(f.Index("daily_spend_summaries_by_date"), dailySpendSummary.Date),
		).In(
			f.If(f.IsNonEmpty(f.Var("dssMatch")),
				f.Update(
					f.Select("ref", f.Get(f.Var("dssMatch"))), f.Obj{"data": dailySpendSummary},
				),
				f.Create(f.Collection("DailySpendSummaries"), f.Obj{
					"data": dailySpendSummary,
				}),
			),
		),
	)

	if err != nil {
		return DailySpendSummary{}, nil, errors.Wrapf(err, "Error creating/updating daily spend summary, %v", err)
	}

	// TODO: I think I may need to do more here to craft the text
	return dailySpendSummary, txns, nil
}

func (s *SpendSummariesService) getToDateMonthlySpend(date string) (float32, error) {
	month, err := timeutil.GetMonthFromDateString(date)
	if err != nil {
		return 0, errors.Wrapf(err, "Failed to parse date when creating dss, err=%v", err)
	}

	readResp, err := s.dbClient.Query(
		f.Select("data",
			f.Reduce(f.Lambda(f.Arr{"acc", "val"}, f.Add(f.Var("acc"), f.Var("val"))),
				0,
				f.Map(f.Paginate(f.MatchTerm(f.Index("daily_spend_summaries_by_month"), month)),
					f.Lambda("ref", f.Select(f.Arr{"data", "spendDiff"}, f.Get(f.Var("ref"))))),
			),
		),
	)
	var runningSpendDiff []float32
	if err = readResp.Get(&runningSpendDiff); err != nil {
		return 0, errors.Wrapf(err, "error reading spend diff from response, %v", err)
	}
	return runningSpendDiff[0], nil
}
