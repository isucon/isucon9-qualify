package scenario

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
)

func check(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	user3 := asset.GetRandomBuyer()

	// 間違ったパスワードでログインができないことをチェックする
	// これがないとパスワードチェックを外して常にログイン成功させるチートが可能になる
	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
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

		var s1, s2 *session.Session
		var err error

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = irregularSellAndBuy(ctx, s1, s2, user3)
			if err != nil {
				critical.Add(err)
			}

			BuyerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
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

		var s1, s2 *session.Session
		var err error

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			// bumpは投稿した直後だとできないので必ず新しいユーザーでやる
			user1 := asset.GetRandomActiveSeller()
			s1, err = loginedSession(ctx, user1)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = bumpAndNewItemsWithLoginedSession(ctx, s1, s2)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

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

		var s1, s2 *session.Session
		var err error
		var targetItemID int64
		var price int

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = newCategoryItemsWithLoginedSession(ctx, s1)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

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

		var s1, s2 *session.Session
		var err error
		var price int
		var targetItemID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = transactionEvidence(ctx, s1)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// ユーザーページを見る
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error
		var price int
		var targetItemID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			// active seller ユーザページ全件確認
			err = loadUserItemsAndItems(ctx, s2, s1.UserID, 2)
			if err != nil {
				critical.Add(err)
				goto Final
			}
			// no active seller ユーザページ確認
			err = loadUserItemsAndItems(ctx, s1, s2.UserID, 0)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// 商品ページをいくつか見る
	// TODO: 初期データをもう少し詰めてから実装する

	// 出品した商品を編集する（100円を110円とかにする）
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error
		var price int
		var targetItemID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = itemEditNewItemWithLoginedSession(ctx, s1, targetItemID, price+10)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, price+10)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
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

		var s1, s2 *session.Session
		var err error
		var price int
		var targetItemID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = loadTransactionEvidence(ctx, s1)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
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
