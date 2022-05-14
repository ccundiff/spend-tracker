package spendsummaries

import (
	"fmt"

	"github.com/akrennmair/slice"
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
	txns = slice.Filter(txns, func(txn transactions.Transaction) bool {
		return !txn.Excluded
	})
	if err != nil {
		return DailySpendSummary{}, nil, errors.Wrapf(err, "Failed to pull txn when creating daily spend summaries, err=%v", err)
	}
	spendTotal := float32(0)
	for _, txn := range txns {
		fmt.Printf("Txn %v=\n", txn)
		spendTotal += txn.Amount
	}

	month, err := timeutil.GetMonthFromDateString(date)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}

	dailySpendSummary := DailySpendSummary{
		Date:       date,
		Month:      month,
		SpendGoal:  user.DailySpendGoal,
		UserId:     user.Id,
		TotalSpend: spendTotal,
		SpendDiff:  float32(user.DailySpendGoal) - spendTotal,
	}

	// TODO: need to account for user id here as well....
	//
	//

	err = s.createOrUpdateDss(dailySpendSummary)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}

	monthlySpendToDate, err := s.getToDateMonthlySpend(date)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}
	dailySpendSummary.ToDateMonthSpend = &monthlySpendToDate
	fmt.Printf("\nmonthly spend what %v=\n", *dailySpendSummary.ToDateMonthSpend)

	// TODO: do the running spend total in fauna after I write the dss so I don't have to use 2 writes her
	err = s.createOrUpdateDss(dailySpendSummary)
	if err != nil {
		return DailySpendSummary{}, nil, err
	}

	if err != nil {
		return DailySpendSummary{}, nil, errors.Wrapf(err, "Error creating/updating daily spend summary, %v", err)
	}

	// TODO: I think I may need to do more here to craft the text
	return dailySpendSummary, txns, nil
}

func (s *SpendSummariesService) createOrUpdateDss(dss DailySpendSummary) error {
	_, err := s.dbClient.Query(
		f.Let().Bind(
			"dssMatch", f.MatchTerm(f.Index("daily_spend_summaries_by_date"), dss.Date),
		).In(
			f.If(f.IsNonEmpty(f.Var("dssMatch")),
				f.Update(
					f.Select("ref", f.Get(f.Var("dssMatch"))), f.Obj{"data": dss},
				),
				f.Create(f.Collection("DailySpendSummaries"), f.Obj{
					"data": dss,
				}),
			),
		),
	)

	return errors.Wrapf(err, "Error in createOrUpdateDss, err=%v", err)
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
					f.Lambda("ref", f.Select(f.Arr{"data", "totalSpend"}, f.Get(f.Var("ref")))),
				),
			),
		))
	var runningSpendTotal []float32
	if err = readResp.Get(&runningSpendTotal); err != nil {
		return 0, errors.Wrapf(err, "error reading spend diff from response, %v", err)
	}
	return runningSpendTotal[0], nil
}
