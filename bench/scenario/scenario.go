package scenario

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
)

func Initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) *fails.Critical {
	critical := fails.NewCritical()

	_, err := initialize(ctx, paymentServiceURL, shipmentServiceURL)
	if err != nil {
		critical.Add(err)
	}

	return critical
}

func Verify(ctx context.Context) *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sellAndBuy(ctx, user1, user2)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		s2, err := LoginedSession(ctx, user2)
		if err != nil {
			critical.Add(err)
			return
		}

		err = bumpAndNewItemsWithLoginedSession(ctx, s1, s2)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		err = newCategoryItemsWithLoginedSession(ctx, s1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		err := itemEdit(ctx, user1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		err = transactionEvidenceWithLoginedSession(ctx, s1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		user1 := asset.GetRandomUser()
		user2 := asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		err = userItemsAndItemWithLoginedSession(ctx, s1, user2.ID)
		if err != nil {
			critical.Add(err)
		}
	}()

	user3 := asset.GetRandomUser()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularLoginWrongPassword(ctx, user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularSellAndBuy(ctx, user2, user1, user3)
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
	closed := make(chan struct{})

	user1, user2, user3 := asset.GetRandomUser(), asset.GetRandomUser(), asset.GetRandomUser()

	// 間違ったパスワードでログインができないことをチェックする
	// これがないとパスワードチェックを外して常にログイン成功させるチートが可能になる
	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			err := irregularLoginWrongPassword(ctx, user3)
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

	// エラー処理が除かれていないかの確認
	// ここだけsellとbuyの間に他の処理がない
	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			err := irregularSellAndBuy(ctx, user2, user1, user3)
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

	// 以下の関数はすべてsellとbuyの間に他の処理を挟む
	// 今回の問題は決済総額がスコアになるのでMySQLを守るためにGETの速度を落とすチートが可能
	// それを防ぐためにsellしたあとに他のエンドポイントにリクエストを飛ばして完了してからbuyされる
	// シナリオとしてはGETで色んなページを見てから初めて購入に結びつくという動きをするのは自然
	// 最適化が難しいエンドポイントの速度をわざと落として、最適化が簡単なエンドポイントに負荷を偏らせるチートを防ぐために
	// すべてのシナリオはチャネルを使って一定時間より早く再実行はしないようにする
	// 理論上そのエンドポイントを高速化することで出せるスコアに上限が出るので、他のエンドポイントを最適化する必要性が出る

	// bumpしてから新着をチェックする
	// TODO: 新着はbumpが新着に出ていることを確認してから、初期データを後ろの方までいい感じに遡りたい
	// TODO: 速度が上がるとbumpしたものが新着に無くなる可能性があるので、created_at的になければ更に遡るようにする
	wg.Add(1)
	go func() {
		defer wg.Done()
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		s2, err := LoginedSession(ctx, user2)
		if err != nil {
			critical.Add(err)
			return
		}

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			err := bumpAndNewItemsWithLoginedSession(ctx, s1, s2)
			if err != nil {
				critical.Add(err)

				goto Final
			}

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// カテゴリ新着をある程度見る
	// TODO: 初期データを後ろの方までいい感じに遡りたい
	wg.Add(1)
	go func() {
		defer wg.Done()
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		s2, err := LoginedSession(ctx, user2)
		if err != nil {
			critical.Add(err)
			return
		}

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			targetItemID, err := sell(ctx, s1, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = newCategoryItemsWithLoginedSession(ctx, s1)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// 取引一覧をある程度見る
	// TODO: 初期データを後ろの方までいい感じに遡りたい
	wg.Add(1)
	go func() {
		defer wg.Done()
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		s2, err := LoginedSession(ctx, user2)
		if err != nil {
			critical.Add(err)
			return
		}

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			targetItemID, err := sell(ctx, s1, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = transactionEvidenceWithLoginedSession(ctx, s1)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// ユーザーページをある程度見る
	// TODO: 初期データを後ろの方までいい感じに遡りたい
	// TODO: ユーザーをランダムにしたい
	wg.Add(1)
	go func() {
		defer wg.Done()
		user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()

		s1, err := LoginedSession(ctx, user1)
		if err != nil {
			critical.Add(err)
			return
		}

		s2, err := LoginedSession(ctx, user2)
		if err != nil {
			critical.Add(err)
			return
		}

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			targetItemID, err := sell(ctx, s1, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = userItemsAndItemWithLoginedSession(ctx, s1, s2.UserID)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// 商品ページをいくつか見る

	// 出品した商品を編集する（100円を110円とかにする）

	go func() {
		wg.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-ctx.Done():
	}
}

func load(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			user1, user2 := asset.GetRandomUser(), asset.GetRandomUser()
			s1, err := LoginedSession(ctx, user1)
			if err != nil {
				critical.Add(err)
				return
			}

			s2, err := LoginedSession(ctx, user2)
			if err != nil {
				critical.Add(err)
				return
			}

		L:
			for j := 0; j < 10; j++ {
				ch := time.After(3 * time.Second)

				err := loadSellNewCategoryBuyWithLoginedSession(ctx, s1, s2)
				if err != nil {
					critical.Add(err)

					goto Final
				}

				err = loadSellNewCategoryBuyWithLoginedSession(ctx, s2, s1)
				if err != nil {
					critical.Add(err)

					goto Final
				}

			Final:
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

			s1, err := LoginedSession(ctx, user1)
			if err != nil {
				critical.Add(err)
				return
			}

			s2, err := LoginedSession(ctx, user2)
			if err != nil {
				critical.Add(err)
				return
			}

		L:
			for j := 0; j < 10; j++ {
				ch := time.After(3 * time.Second)

				targetItemID, err := sell(ctx, s1, 100)
				if err != nil {
					critical.Add(err)

					goto Final
				}

				err = userItemsAndItemWithLoginedSession(ctx, s1, user2.ID)
				if err != nil {
					critical.Add(err)

					goto Final
				}

				err = buyComplete(ctx, s1, s2, targetItemID, 100)
				if err != nil {
					critical.Add(err)

					goto Final
				}

			Final:
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

func Campaign(critical *fails.Critical) {}

func FinalCheck(critical *fails.Critical) {}

var (
	sShipment *server.ServerShipment
	sPayment  *server.ServerPayment
)
