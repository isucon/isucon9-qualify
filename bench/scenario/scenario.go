package scenario

import (
	"sync"

	"github.com/isucon/isucon9-qualify/bench/fails"
)

func Initialize() {
}

func Verify() *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sellAndBuy()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := bump()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularWrongPassword()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularSellWrongCSRFToken()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularSellWrongPrice()
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}
