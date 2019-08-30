package scenario

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	ExecutionSeconds = 60
)

func Initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (bool, *fails.Critical) {
	critical := fails.NewCritical()

	// initializeだけタイムアウトを別に設定
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	isCampaign, err := initialize(ctx, paymentServiceURL, shipmentServiceURL)
	if err != nil {
		critical.Add(err)
	}

	return isCampaign, critical
}

func Verify(ctx context.Context) *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		targetItemID, err := sell(ctx, s1, 100)
		if err != nil {
			critical.Add(err)
			return
		}

		err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, 100)
		if err != nil {
			critical.Add(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		err = bumpAndNewItemsWithLoginedSession(ctx, s1, s2)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		err = newCategoryItemsWithLoginedSession(ctx, s1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		targetItemID := asset.GetUserItemsFirst(s1.UserID)

		err = itemEditWithLoginedSession(ctx, s1, targetItemID, 110)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		err = transactionEvidence(ctx, s1)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		s2, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s2)

		err = userItemsAndItem(ctx, s1, s2.UserID)
		if err != nil {
			critical.Add(err)
		}

		// active sellerの全件確認
		err = countUserItems(ctx, s2, s2.UserID)
		if err != nil {
			critical.Add(err)
		}
		// active sellerではないユーザも確認。0件でも問題ない
		err = countUserItems(ctx, s1, s2.UserID)
		if err != nil {
			critical.Add(err)
		}
	}()

	user3 := asset.GetRandomBuyer()

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
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		err = irregularSellAndBuy(ctx, s1, s2, user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}

func Validation(ctx context.Context, isCampaign bool, critical *fails.Critical) {
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

	if isCampaign {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Print("=== enable campaign ===")
			Campaign(ctx, critical)
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
			err = countUserItems(ctx, s2, s1.UserID)
			if err != nil {
				critical.Add(err)
				goto Final
			}
			// no active seller ユーザページ確認
			err = countUserItems(ctx, s1, s2.UserID)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = userItemsAndItem(ctx, s2, s1.UserID)
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

func load(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	// load scenario #1
	// カテゴリを少しみてbuy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var categories []asset.AppCategory
		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				categories = asset.GetRootCategories()
				for _, category := range categories {
					err = newCategoryItemsAndItems(ctx, s1, category.ID, 20, 15)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				err = loadSellNewCategoryBuyWithLoginedSession(ctx, s1, s2, price)
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
	}

	// load scenario #2
	// どちらかというとカテゴリを中心にみていく
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var targetItemID int64
			var item *session.ItemDetail
		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

				s2, err = activeSellerSession(ctx)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				s1, err = buyerSession(ctx)
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

				item, err = s1.Item(ctx, targetItemID)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = newCategoryItemsAndItems(ctx, s2, item.Category.ParentID, 20, 5)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = transactionEvidence(ctx, s1)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				// ここは厳密なcheckをしない
				err = buyComplete(ctx, s1, s2, targetItemID, price)
				if err != nil {
					critical.Add(err)

					goto Final
				}

			Final:
				ActiveSellerPool.Enqueue(s1)
				BuyerPool.Enqueue(s2)

				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	// load scenario #3
	// どちらかというとuserを中心にみていく
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

				s1, err := buyerSession(ctx)
				if err != nil {
					critical.Add(err)
					return
				}

				s2, err := activeSellerSession(ctx)
				if err != nil {
					critical.Add(err)
					return
				}

				s3, err := buyerSession(ctx)
				if err != nil {
					critical.Add(err)
					return
				}

				price := priceStoreCache.Get()

				targetItemID, err := sell(ctx, s1, price)
				if err != nil {
					critical.Add(err)

					goto Final
				}

				// ユーザのページを全部みる。
				// activeユーザ3ページ
				err = countUserItems(ctx, s1, s2.UserID)
				if err != nil {
					critical.Add(err)
					goto Final
				}
				// 商品数がすくないところもみにいく
				// indexつけるだけで速くなる
				for l := 0; l < 4; l++ {
					err = countUserItems(ctx, s1, s3.UserID)
					if err != nil {
						critical.Add(err)
						goto Final
					}
					err = countUserItems(ctx, s3, s1.UserID)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				err = userItemsAndItem(ctx, s1, s2.UserID)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
				if err != nil {
					critical.Add(err)

					goto Final
				}

			Final:
				ActiveSellerPool.Enqueue(s2)
				BuyerPool.Enqueue(s1)

				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	// load scenario #4
	// NewItemみてbuy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

				s1, err := activeSellerSession(ctx)
				if err != nil {
					critical.Add(err)
					return
				}

				s2, err := buyerSession(ctx)
				if err != nil {
					critical.Add(err)
					return
				}

				price := priceStoreCache.Get()

				targetItemID, err := sell(ctx, s1, price)
				if err != nil {
					critical.Add(err)

					goto Final
				}

				err = newItemsAndItems(ctx, s2, 30, 50)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
				if err != nil {
					critical.Add(err)
					goto Final
				}

			Final:
				ActiveSellerPool.Enqueue(s1)
				BuyerPool.Enqueue(s2)

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

func FinalCheck(ctx context.Context, critical *fails.Critical) int64 {
	reports := sPayment.GetReports()

	s1, err := session.NewSession()
	if err != nil {
		critical.Add(err)

		return 0
	}

	tes, err := s1.Reports(ctx)
	if err != nil {
		critical.Add(err)

		return 0
	}

	var score int64

	for _, te := range tes {
		report, ok := reports[te.ItemID]
		if !ok {
			critical.Add(failure.New(fails.ErrApplication, failure.Messagef("購入実績がありません transaction_evidence_id: %d; item_id: %d", te.ID, te.ItemID)))
			continue
		}

		if report.Price != te.ItemPrice {
			critical.Add(failure.New(fails.ErrApplication, failure.Messagef("購入実績の価格が異なります transaction_evidence_id: %d; item_id: %d; expected price: %d; reported price: %d", te.ID, te.ItemID, report.Price, te.ItemPrice)))
			continue
		}

		score += int64(report.Price)
		delete(reports, te.ItemID)
	}

	for itemID, report := range reports {
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("購入されたはずなのに記録されていません item_id: %d; expected price: %d", itemID, report.Price)))
	}

	return score
}

var (
	sShipment *server.ServerShipment
	sPayment  *server.ServerPayment
)
