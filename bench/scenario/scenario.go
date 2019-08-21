package scenario

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
)

func Initialize() *fails.Critical {
	critical := fails.NewCritical()

	_, err := initialize("", "")
	if err != nil {
		critical.Add(err)
	}

	return critical
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
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()
		err := bumpAndNewItems(user1, user2)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		err := newCategoryItems(user1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		err := itemEdit(user1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		err := transactionEvidence(user1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		user2 := asset.GetRandomUser()

		s1, err := LoginedSession(user1)
		if err != nil {
			critical.Add(err)
			return
		}

		err = userItemsAndItemWithLoginedSession(s1, user2.ID)
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
		err := irregularSellAndBuy(user2, user1, user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}

func Validation(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		check(ctx, critical)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		load(ctx, critical)
	}()

	go func() {
		wg.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-ctx.Done():
	}
}

func check(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup

	user1, user2, user3 := asset.GetRandomUser(), asset.GetRandomUser(), asset.GetRandomUser()

	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			err := irregularLoginWrongPassword(user3)
			if err != nil {
				critical.Add(err)
			}

			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			err := irregularSellAndBuy(user2, user1, user3)
			if err != nil {
				critical.Add(err)
			}

			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	wg.Wait()
}

func load(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()
			s1, err := LoginedSession(user1)
			if err != nil {
				critical.Add(err)
				return
			}

			s2, err := LoginedSession(user2)
			if err != nil {
				critical.Add(err)
				return
			}

		L:
			for j := 0; j < 10; j++ {
				ch := time.After(3 * time.Second)

				err := loadSellNewCategoryBuyWithLoginedSession(s1, s2)
				if err != nil {
					critical.Add(err)

					goto Last
				}

				err = loadSellNewCategoryBuyWithLoginedSession(s2, s1)
				if err != nil {
					critical.Add(err)

					goto Last
				}

			Last:
				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user1 := asset.GetRandomUser()
			user2 := asset.GetRandomUser()

			s1, err := LoginedSession(user1)
			if err != nil {
				critical.Add(err)
				return
			}

			s2, err := LoginedSession(user2)
			if err != nil {
				critical.Add(err)
				return
			}

		L:
			for j := 0; j < 10; j++ {
				ch := time.After(3 * time.Second)

				targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
				if err != nil {
					critical.Add(err)

					goto Last
				}

				err = userItemsAndItemWithLoginedSession(s1, user2.ID)
				if err != nil {
					critical.Add(err)

					goto Last
				}

				err = buyComplete(s1, s2, targetItemID)
				if err != nil {
					critical.Add(err)

					goto Last
				}

			Last:
				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-ctx.Done():
	}
}

func FinalCheck(critical *fails.Critical) {}

var (
	sShipment *server.ServerShipment
	sPayment  *server.ServerPayment
)
