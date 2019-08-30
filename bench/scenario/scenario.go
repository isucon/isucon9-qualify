package scenario

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
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

		targetItemID, fileName, err := sellForFileName(ctx, s1, 100)
		if err != nil {
			critical.Add(err)
			return
		}

		f, err := os.Open(fileName)
		if err != nil {
			critical.Add(failure.Wrap(err, failure.Message("ベンチマーカー内部のファイルを開くことに失敗しました")))
			return
		}

		h := md5.New()
		_, err = io.Copy(h, f)
		if err != nil {
			critical.Add(failure.Wrap(err, failure.Message("ベンチマーカー内部のファイルのmd5値を取ることに失敗しました")))
			return
		}

		expectedMD5Str := fmt.Sprintf("%x", h.Sum(nil))

		item, err := s1.Item(ctx, targetItemID)
		if err != nil {
			critical.Add(err)
			return
		}

		md5Str, err := s1.DownloadItemImageURL(ctx, item.ImageURL)
		if err != nil {
			critical.Add(err)
			return
		}

		if expectedMD5Str != md5Str {
			critical.Add(failure.New(fails.ErrApplication, failure.Messagef("%sの画像のmd5値が間違っています expected: %s; actual: %s", item.ImageURL, expectedMD5Str, md5Str)))
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

		// buyer の全件確認 (self)
		err = userItemsAllAndItems(ctx, s1, s1.UserID, 0)
		if err != nil {
			critical.Add(err)
			return
		}

		// active sellerの全件確認(self)
		err = userItemsAllAndItems(ctx, s2, s2.UserID, 5)
		if err != nil {
			critical.Add(err)
			return
		}

		// active sellerではないユーザも確認。0件でも問題ない
		userIDs := asset.GetRandomBuyerIDs(10)
		for _, userID := range userIDs {
			err = userItemsAllAndItems(ctx, s1, userID, 0)
			if err != nil {
				critical.Add(err)
			}
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

		// active sellerの全件確認(random)
		userIDs := asset.GetRandomActiveSellerIDs(20)
		for _, userID := range userIDs {
			err = userItemsAllAndItems(ctx, s1, userID, 2)
			if err != nil {
				critical.Add(err)
				return
			}
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
