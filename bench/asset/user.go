package asset

type AppUser struct {
	ID          int64  `json:"id"`
	AccountName string `json:"account_name"`
	Password    string `json:"-"`
	Address     string `json:"address,omitempty"`
}

var users []AppUser

func init() {
	users = make([]AppUser, 0, 100)
	users = []AppUser{
		AppUser{
			AccountName: "aaa",
			Address:     "aaa",
			Password:    "aaa",
		},
		AppUser{
			AccountName: "bbb",
			Address:     "bbb",
			Password:    "bbb",
		},
	}
}

func (u1 *AppUser) Equal(u2 *AppUser) bool {
	return u1.AccountName == u2.AccountName && u1.Address == u2.Address
}

func GetRandomUserPair() (AppUser, AppUser) {
	return users[0], users[1]
}

func GetRandomUser() AppUser {
	return users[1]
}
