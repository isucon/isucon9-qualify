package scenario

import (
	"context"
	"log"
	"sync"
	"time"

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

func Validation(ctx context.Context, isCampaign bool, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		Check(ctx, critical)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		Load(ctx, critical)
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
