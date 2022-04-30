package users

import f "github.com/fauna/faunadb-go/v4/faunadb"

type User struct {
	Id         f.RefV `fauna:"ref"`
	Name        string `fauna:"name"`
	AccessToken string `fauna:"accessToken"`
	ItemId      string `fauna:"itemId"`
	DailySpendGoal int `fauna:"dailySpendGoal"`
}
