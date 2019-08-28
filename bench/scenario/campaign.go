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

func Campaign(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	go func() {
	L:
		for j := 0; j < 10; j++ {
			ch := time.After(8 * time.Second)

			popularListing(ctx, critical)

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
func popularListing(ctx context.Context, critical *fails.Critical) {
	// buyerが足りない分を準備しておく
	if l := BuyerPool.Len(); l < 50 {
		count := 60 - l

		var wg sync.WaitGroup

		for i := 0; i < count/5; i++ {
			for j := 0; j < 5; j++ {
				wg.Add(1)

				go func() {
					defer wg.Done()

					user1 := asset.GetRandomBuyer()
					s, err := loginedSession(ctx, user1)
					if err != nil {
						critical.Add(err)
						return
					}
					BuyerPool.Enqueue(s)
				}()
			}
			// 一気にログインするとアプリケーションがしんどいのでほどほどにする
			<-time.After(100 * time.Millisecond)
		}

		wg.Wait()
	}

	// 真のbuyerが入るチャネル。複数来たらエラーにする
	buyerCh := make(chan *session.Session, 0)
	defer func() {
		// closeしないとgoroutineリークする
		close(buyerCh)
	}()

	popular, err := buyerSession(ctx)
	if err != nil {
		critical.Add(err)
		return
	}

	price := 1000

	targetItemID, err := sell(ctx, popular, price)
	if err != nil {
		critical.Add(err)
		return
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			token := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

			s2, err := buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				return
			}

			transactionEvidenceID, err := s2.BuyWithMayFail(ctx, targetItemID, token)
			if err != nil {
				critical.Add(err)
				return
			}

			if transactionEvidenceID != 0 {
				// 0でないなら真のbuyer
				// 関数を終わらせないとデッドロックするのでgoroutineを使う
				go func() {
					buyerCh <- s2
				}()
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
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("商品 (item_id: %d) に対して全ユーザーが購入に失敗しました", targetItemID)))
		return
	}

	go func() {
		for s := range buyerCh {
			// TODO: ここはクリティカル扱いにして発生していたら大幅減点にしたい
			// buyerが複数人いるとここのコードが動く
			critical.Add(failure.New(fails.ErrApplication, failure.Messagef("購入済み商品 (item_id: %d) に対して他のユーザー (user_id: %d) が購入できています", targetItemID, s.UserID)))
		}
	}()

	reserveID, apath, err := popular.Ship(ctx, targetItemID)
	if err != nil {
		critical.Add(err)
		return
	}

	md5Str, err := popular.DownloadQRURL(ctx, apath)
	if err != nil {
		critical.Add(err)
		return
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		critical.Add(failure.New(fails.ErrApplication, failure.Message("QRコードの画像に誤りがあります")))
		return
	}

	err = popular.ShipDone(ctx, targetItemID)
	if err != nil {
		critical.Add(err)
		return
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		critical.Add(failure.New(fails.ErrApplication, failure.Message("配送予約IDに誤りがあります")))
		return
	}

	err = buyer.Complete(ctx, targetItemID)
	if err != nil {
		critical.Add(err)
		return
	}

	return
}
