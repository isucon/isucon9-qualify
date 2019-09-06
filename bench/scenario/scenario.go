package scenario

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	ExecutionSeconds = 60
)

func Initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (int, string) {
	// initializeだけタイムアウトを別に設定
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	campaign, language, err := initialize(ctx, paymentServiceURL, shipmentServiceURL)
	if err != nil {
		fails.ErrorsForCheck.Add(err)
	}

	return campaign, language
}

func Validation(ctx context.Context, campaign int) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		Check(ctx)
	}()

	/*
		キャンペーンの還元率(の設定)で負荷が変わる
		還元率の設定, 負荷, 人気者出品
		0, 2, なし
		1, 3, あり
		2, 4, あり
		3, 5, あり
		4, 6, あり
	*/
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Print("- Start Load worker 1")
		Load(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-time.After(100 * time.Millisecond)
		log.Print("- Start Load worker 2")
		Load(ctx)
	}()

	if campaign > 0 {
		log.Printf("=== enable campaign rate setting => %d ===", campaign)
		for i := 0; i < campaign; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-time.After(time.Duration((i+2)*100) * time.Millisecond)
				log.Printf("- Start Load worker %d", i+3)
				Load(ctx)
			}(i)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			Campaign(ctx)
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

func FinalCheck(ctx context.Context) int64 {
	reports := sPayment.GetReports()

	s1, err := session.NewSession()
	if err != nil {
		fails.ErrorsForFinal.Add(err)

		return 0
	}

	tes, err := s1.Reports(ctx)
	if err != nil {
		fails.ErrorsForFinal.Add(err)

		return 0
	}

	var score int64

	for _, te := range tes {
		report, ok := reports[te.ItemID]
		if !ok {
			fails.ErrorsForFinal.Add(failure.New(fails.ErrApplication, failure.Messagef("購入実績がありません transaction_evidence_id: %d; item_id: %d", te.ID, te.ItemID)))
			continue
		}

		delete(reports, te.ItemID)

		if report.Price != te.ItemPrice {
			fails.ErrorsForFinal.Add(failure.New(fails.ErrApplication, failure.Messagef("購入実績の価格が異なります transaction_evidence_id: %d; item_id: %d; expected price: %d; reported price: %d", te.ID, te.ItemID, report.Price, te.ItemPrice)))
			continue
		}

		// statusのチェックはこちらからコネクションを切断したケースでずれる可能性がある
		// とりあえずチェックせず、こちらがdoneだと認めたケースだけで加点する

		if report.Status == asset.TransactionEvidenceStatusDone {
			// doneの時だけが売り上げとして認められる
			score += int64(report.Price)
		}
	}

	for itemID, report := range reports {
		fails.ErrorsForFinal.Add(failure.New(fails.ErrApplication, failure.Messagef("購入されたはずなのに記録されていません item_id: %d; expected price: %d", itemID, report.Price)))
	}

	return score
}
