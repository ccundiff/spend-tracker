package users

import (
	f "github.com/fauna/faunadb-go/v4/faunadb"
)

type UserService struct {
	dbClient  *f.FaunaClient
}

func NewUsersService(dbClient *f.FaunaClient) *UserService {
	return &UserService{dbClient: dbClient}
}

func (u *UserService) GetUser() (User, error) {
	userRes, err := u.dbClient.Query(
		f.Let().Bind("user", f.Get(f.Ref(f.Collection("Users"), "326065743220179012"))).In(
			f.Obj{
				"ref":         f.Select("ref", f.Var("user")),
				"name":        f.Select(f.Arr{"data", "name"}, f.Var("user")),
				"accessToken": f.Select(f.Arr{"data", "accessToken"}, f.Var("user")),
				"itemId":      f.Select(f.Arr{"data", "itemId"}, f.Var("user")),
				"dailySpendGoal": f.Select(f.Arr{"data", "dailySpendGoal"}, f.Var("user")),
			},
		),
	)

	if err != nil {
		return User{}, err
	}

	var user User

	if err := userRes.Get(&user); err != nil {
		return User{}, err
	}

	return user, nil
}
