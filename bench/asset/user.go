package asset

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
)

type AppUser struct {
	ID           int64  `json:"id"`
	AccountName  string `json:"account_name"`
	Password     string `json:"plain_passwd"`
	Address      string `json:"address,omitempty"`
	NumSellItems int    `json:"num_sell_items"`
}

type AppItem struct {
	ID          int64  `json:"id"`
	SellerID    int64  `json:"seller_id"`
	BuyerID     int64  `json:"buyer_id"`
	Status      string `json:"status"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
}

var (
	users     []AppUser
	Items     map[int64][]AppItem
	indexUser int32
)

func init() {
	users = make([]AppUser, 0, 100)
	Items = make(map[int64][]AppItem)

	f, err := os.Open("initial-data/result/users_json.txt")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	user := &AppUser{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), user)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, *user)
	}
	f.Close()

	f, err = os.Open("initial-data/result/items_json.txt")
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(f)
	item := &AppItem{}

	for scanner.Scan() {
		err := json.Unmarshal([]byte(scanner.Text()), item)
		if err != nil {
			log.Fatal(err)
		}
		Items[item.SellerID] = append(Items[item.SellerID], *item)
	}
	f.Close()

	rand.Shuffle(len(users), func(i, j int) { users[i], users[j] = users[j], users[i] })
}

func (u1 *AppUser) Equal(u2 *AppUser) bool {
	return u1.AccountName == u2.AccountName && u1.Address == u2.Address
}

func GetRandomUser() AppUser {
	// 全部使い切ったらpanicするので十分なユーザー数を用意しておく
	return users[len(users)-int(atomic.AddInt32(&indexUser, 1))]
}
