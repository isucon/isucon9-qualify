package scenario

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func Campaign(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	// buyer用のセッションを増やしておく
	// 500ユーザーを追加したら止まる
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		L:
			for j := 0; j < 50; j++ {
				ch := time.After(100 * time.Millisecond)

				user1 := asset.GetRandomBuyer()
				s, err := loginedSession(ctx, user1)
				if err != nil {
					// ログインに失敗しまくるとプールに溜まらないので一気に購入できなくなる
					// その場合は失敗件数が多いという理由で失格にする
					critical.Add(err)
					goto Final
				}
				BuyerPool.Enqueue(s)

			Final:
				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// ログインユーザーがある程度溜まらないと実施できないので少し待つ
		// 8s毎に実行されるので60sだと最大で5回実行される
		<-time.After(13 * time.Second)

	L:
		for j := 0; j < (ExecutionSeconds-13)/8; j++ {
			ch := time.After(8 * time.Second)

			isIncrease := popularListing(ctx, critical, 80+j*20, 1000+j*100)

			if isIncrease {
				// 商品単価を上げる
				log.Print("=== succeed to popular listing ===")

				priceStoreCache.Add(20)

				// 次の人気者出品に備えてログインユーザーのpoolを増やしておく
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()

					L:
						for j := 0; j < 20; j++ {
							ch := time.After(100 * time.Millisecond)

							user1 := asset.GetRandomBuyer()
							s, err := loginedSession(ctx, user1)
							if err != nil {
								// ログインに失敗しまくるとプールに溜まらないので一気に購入できなくなる
								// その場合は失敗件数が多いという理由で失格にする
								critical.Add(err)
								goto Final
							}
							BuyerPool.Enqueue(s)

						Final:
							select {
							case <-ch:
							case <-ctx.Done():
								break L
							}
						}
					}()
				}
			}

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

// popularListing is 人気者出品
// 人気者が高額の出品を行う。高額だが出品した瞬間に大量の人が購入しようとしてくる。もちろん購入できるのは一人だけ。
func popularListing(ctx context.Context, critical *fails.Critical, num int, price int) (isIncrease bool) {
	// buyerが足りない場合はログインを意図的に遅くしている可能性があるのでペナルティとして実行しない
	l := BuyerPool.Len()
	if l < num+10 {
		log.Printf("login user insufficient (count: %d)", l)
		return false
	}

	// 真のbuyerが入るチャネル。複数来たらエラーにする
	buyerCh := make(chan *session.Session, 1)

	popular, err := buyerSession(ctx)
	if err != nil {
		critical.Add(err)
		return false
	}

	// 人気者出品だけはだれが買うかわからないので、カテゴリ指定なし出品
	targetItem, err := sell(ctx, popular, price)
	if err != nil {
		critical.Add(err)
		return false
	}

	var wg sync.WaitGroup
	var errCnt int32

	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			token := sPayment.ForceSet(CorrectCardNumber, targetItem.ID, price)

			s2, err := buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				atomic.AddInt32(&errCnt, 1)
				return
			}

			transactionEvidenceID, err := s2.BuyWithMayFail(ctx, targetItem.ID, token)
			if err != nil {
				critical.Add(err)
				atomic.AddInt32(&errCnt, 1)
				return
			}

			if transactionEvidenceID != 0 {
				// 0でないなら真のbuyer
				buyerCh <- s2
			} else {
				// buyerでないならもう使わないので戻す
				BuyerPool.Enqueue(s2)
			}
		}()
	}

	closed := make(chan struct{})
	go func() {
		wg.Wait()
		close(closed)
	}()

	var buyer *session.Session

	select {
	case buyer = <-buyerCh:
	case <-closed:
		// 全goroutineが終了したのにbuyerがいない場合は全員が購入に失敗している
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("商品 (item_id: %d) に対して全ユーザーが購入に失敗しました", targetItem.ID)))
		return false
	}

	defer func() {
		// 終わったら戻しておく
		BuyerPool.Enqueue(buyer)
	}()

	go func() {
	L:
		for {
			select {
			case s := <-buyerCh:
				// buyerが複数人いるとここのコードが動く
				critical.Add(failure.New(fails.ErrCritical, failure.Messagef("購入済み商品 (item_id: %d) に対して他のユーザー (user_id: %d) が購入できています", targetItem.ID, s.UserID)))
			case <-closed:
				break L
			}
		}
	}()

	reserveID, apath, err := popular.Ship(ctx, targetItem.ID)
	if err != nil {
		critical.Add(err)
		return false
	}

	md5Str, err := popular.DownloadQRURL(ctx, apath)
	if err != nil {
		critical.Add(err)
		return false
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("QRコードの画像に誤りがあります (item_id: %d, reserve_id: %s)", targetItem.ID, reserveID)))
		return false
	}

	err = popular.ShipDone(ctx, targetItem.ID)
	if err != nil {
		critical.Add(err)
		return false
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("配送予約IDに誤りがあります (item_id: %d, reserve_id: %s)", targetItem.ID, reserveID)))
		return false
	}

	err = buyer.Complete(ctx, targetItem.ID)
	if err != nil {
		critical.Add(err)
		return false
	}

	if atomic.LoadInt32(&errCnt) > 2 {
		// エラーが一定数を超えていたら単価は上がらない
		return false
	}

	return true
}
