package scenario

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
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

	wg.Add(1)
	go func() {
		defer wg.Done()

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

		targetItemID, err := sell(ctx, s1, 100)
		if err != nil {
			critical.Add(err)
			return
		}

		err = buyComplete(ctx, s1, s2, targetItemID, 100)
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

		s2, err := buyerSession(ctx)
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
		s1, err := buyerSession(ctx)
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
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}

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

		err = transactionEvidenceWithLoginedSession(ctx, s1)
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

		s2, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}

		err = userItemsAndItemWithLoginedSession(ctx, s1, s2.UserID)
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

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}

		err = irregularSellAndBuy(ctx, s1, s2, user3)
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

	user3 := asset.GetRandomBuyer()

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

			s1, err := buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				return
			}

			s2, err := buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				return
			}

			err = irregularSellAndBuy(ctx, s1, s2, user3)
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
	// TODO: 商品ページも見るのは蛇足では
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			targetItemID, err := sell(ctx, s1, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = userItemsAndItemWithLoginedSession(ctx, s2, s1.UserID)
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
	// TODO: 初期データをもう少し詰めてから実装する

	// 出品した商品を編集する（100円を110円とかにする）
	wg.Add(1)
	go func() {
		defer wg.Done()
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

	L:
		for j := 0; j < 10; j++ {
			ch := time.After(5 * time.Second)

			targetItemID, err := sell(ctx, s1, 100)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = itemEditNewItemWithLoginedSession(ctx, s1, targetItemID, 110)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, 110)
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

		L:
			for j := 0; j < 10; j++ {
				ch := time.After(3 * time.Second)

				err := loadSellNewCategoryBuyWithLoginedSession(ctx, s1, s2)
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
			s2, err := activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				return
			}

			s1, err := buyerSession(ctx)
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
