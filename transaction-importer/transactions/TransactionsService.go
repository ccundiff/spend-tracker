package transactions

import (
	"context"
	"fmt"
	"github.com/ccundiff/spend-tracker/transaction-importer/users"
	f "github.com/fauna/faunadb-go/v4/faunadb"
	"github.com/plaid/plaid-go/plaid"
)

type Transaction struct {
	Merchant string  `fauna:"merchant"`
	Amount   float32 `fauna:"amount"`
}

type TransactionsService struct {
	DbClient    *f.FaunaClient
	PlaidClient *plaid.APIClient
}

func NewTransactionsService(dbClient *f.FaunaClient, plaidClient *plaid.APIClient) *TransactionsService {
	return &TransactionsService{
		DbClient:    dbClient,
		PlaidClient: plaidClient,
	}
}

func (t *TransactionsService) ImportTransactions(user users.User, date string) error {
	request := plaid.NewTransactionsGetRequest(
		user.AccessToken,
		date,
		date,
	)

	options := plaid.TransactionsGetRequestOptions{
		Count:  plaid.PtrInt32(500),
		Offset: plaid.PtrInt32(0),
	}

	request.SetOptions(options)

	// todo: limit this to only my credit card accounts or just account for debits
	transactionsResp, _, err := t.PlaidClient.PlaidApi.TransactionsGet(context.TODO()).TransactionsGetRequest(*request).Execute()
	if err != nil {
		plaidErr, err := plaid.ToPlaidError(err)
		println(fmt.Sprintf("unable to pull txns, plaiderr=[%v], err=[%v]", plaidErr, err))
		return err
	}

	createStatements := make([]f.Expr, len(transactionsResp.Transactions))

	for _, txn := range transactionsResp.Transactions {
		createStatements = append(createStatements,
			f.If(
				f.Not(f.Exists(f.MatchTerm(f.Index("transactions_by_plaid_id"), txn.TransactionId))),
				f.Create(f.Collection("Transactions"), f.Obj{
					"data": f.Obj{
						"plaidTransactionId": txn.TransactionId,
						"merchant":           txn.Name,
						"amount":             txn.Amount,
						"user":               user.Id,
						"date":               txn.Date,
					},
				}),
				// todo: do I want to update here?
				f.Null(),
			))
	}

	_, err = t.DbClient.Query(
		f.Do(createStatements),
	)

	if err != nil {
		fmt.Printf("error creating txns and spend summary, err=[%v]", err)
	}
	return err
}

// TODO: need to include user in this
// TODO: setup txns to have the field that they aren't counted in budgeting and filter on that here, add separate method for pulling ignored txns
func (t *TransactionsService) TransactionsForDate(date string) ([]Transaction, error) {
	println(date)
	query := f.Map(
		f.Paginate(f.MatchTerm(f.Index("transactions_by_date"), date)),
		f.Lambda("txnRef",
			f.Let().Bind("txn", f.Get(f.Var("txnRef"))).In(
				f.Obj{
					"merchant": f.Select(f.Arr{"data", "merchant"}, f.Var("txn")),
					"amount":   f.Select(f.Arr{"data", "amount"}, f.Var("txn")),
				},
			),
		),
	)

	txnsRes, err := t.DbClient.Query(query)

	if err != nil {
		return nil, err
	}

	var parsedTxns []Transaction

	if err = txnsRes.At(f.ObjKey("data")).Get(&parsedTxns); err != nil {
		return nil, err
	}

	return parsedTxns, nil
}
