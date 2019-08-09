package scenario

import (
	"sync"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
)

func Initialize() {
}

func Verify() *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sellAndBuy(user1, user2)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.AppUser{
			AccountName: "aaa",
			Address:     "aaa",
			Password:    "aaa",
		}
		user2 := asset.AppUser{
			AccountName: "bbb",
			Address:     "bbb",
			Password:    "bbb",
		}
		// bumpするためにはそのユーザーのItemIDが必要
		err := bump(user1, user2)
		if err != nil {
			critical.Add(err)
		}
	}()

	user3 := asset.GetRandomUser()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularLoginWrongPassword(user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularSell(user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}
