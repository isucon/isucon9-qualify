package asset

import (
	"math/rand"
	"sync/atomic"
)

type AppUser struct {
	ID          int64  `json:"id"`
	AccountName string `json:"account_name"`
	Password    string `json:"-"`
	Address     string `json:"address,omitempty"`
}

var users []AppUser
var index int32

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
		AppUser{
			AccountName: "ccc",
			Address:     "ccc",
			Password:    "ccc",
		},
	}
	rand.Shuffle(len(users), func(i, j int) { users[i], users[j] = users[j], users[i] })
}

func (u1 *AppUser) Equal(u2 *AppUser) bool {
	return u1.AccountName == u2.AccountName && u1.Address == u2.Address
}

func GetRandomUser() AppUser {
	// 全部使い切ったらpanicするので十分なユーザー数を用意しておく
	return users[len(users)-int(atomic.AddInt32(&index, 1))]
}
