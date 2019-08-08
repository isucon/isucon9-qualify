package scenario

import (
	"time"

	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/k0kubun/pp"
)

func SellAndBuy() error {
	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	s2, err := session.NewSession()
	if err != nil {
		return err
	}

	seller, err := s1.Login("aaa", "aaa")
	if err != nil {
		return err
	}
	pp.Println(seller)
	err = s1.SetSettings()
	if err != nil {
		return err
	}

	buyer, err := s2.Login("bbb", "bbb")
	if err != nil {
		return err
	}
	pp.Println(buyer)
	err = s2.SetSettings()
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}
	token, err := s2.PaymentCard("AAAAAAAA", "11")
	if err != nil {
		return err
	}
	err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	aurl, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	s3, err := session.NewSession()
	if err != nil {
		return err
	}
	surl, err := s3.DecodeQRURL(aurl)
	if err != nil {
		return err
	}

	err = s3.ShipmentAccept(surl)
	if err != nil {
		return err
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	<-time.After(6 * time.Second)

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}
