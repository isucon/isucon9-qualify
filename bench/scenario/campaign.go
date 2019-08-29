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
				ch := time.After(200 * time.Millisecond)

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
		<-time.After(7 * time.Second)

	L:
		for j := 0; j < ExecutionSeconds/10; j++ {
			ch := time.After(10 * time.Second)

			isIncrease := popularListing(ctx, critical, 80)

			if isIncrease {
				// goroutineを増やす
				log.Print("increase")
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
func popularListing(ctx context.Context, critical *fails.Critical, num int) (isIncrease bool) {
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

	price := 1000

	targetItemID, err := sell(ctx, popular, price)
	if err != nil {
		critical.Add(err)
		return false
	}

	var wg sync.WaitGroup

	for i := 0; i < num; i++ {
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
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("商品 (item_id: %d) に対して全ユーザーが購入に失敗しました", targetItemID)))
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
				critical.Add(failure.New(fails.ErrCritical, failure.Messagef("購入済み商品 (item_id: %d) に対して他のユーザー (user_id: %d) が購入できています", targetItemID, s.UserID)))
			case <-closed:
				break L
			}
		}
	}()

	reserveID, apath, err := popular.Ship(ctx, targetItemID)
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
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("QRコードの画像に誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID)))
		return false
	}

	err = popular.ShipDone(ctx, targetItemID)
	if err != nil {
		critical.Add(err)
		return false
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		critical.Add(failure.New(fails.ErrApplication, failure.Messagef("配送予約IDに誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID)))
		return false
	}

	err = buyer.Complete(ctx, targetItemID)
	if err != nil {
		critical.Add(err)
		return false
	}

	return true
}
