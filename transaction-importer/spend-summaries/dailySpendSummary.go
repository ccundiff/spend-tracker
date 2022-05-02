package spendsummaries

import f "github.com/fauna/faunadb-go/v4/faunadb"

type DailySpendSummary struct {
	Date             string   `fauna:"date"`
	Month            int      `fauna:"month"`
	SpendGoal        int      `fauna:"spendGoal"`
	UserId           f.RefV   `fauna:"userRef"`
	TotalSpend       float32  `fauna:"totalSpend"`
	SpendDiff        float32  `fauna:"spendDiff"`
	ToDateMonthSpend *float32 `fauna:"toDateMonthSpend",omitempty`
}
