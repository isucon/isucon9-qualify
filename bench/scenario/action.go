package scenario

import (
	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	CorrectCardNumber = "AAAAAAAA"
	FailedCardNumber  = "FA10AAAA"
	IsucariShopID     = "11"

	ErrScenario failure.StringCode = "error scenario"
)

func LoginedSession(user1 asset.AppUser) (*session.Session, error) {
	s1, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	user, err := s1.Login(user1.AccountName, user1.Password)
	if err != nil {
		return nil, err
	}

	if !user1.Equal(user) {
		return nil, failure.New(ErrScenario, failure.Message("ログインが失敗しています"))
	}

	err = s1.SetSettings()
	if err != nil {
		return nil, err
	}

	return s1, nil
}

func buyComplete(s1, s2 *session.Session, targetItemID int64) error {
	token, err := s2.PaymentCard(CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}
	_, err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(ErrScenario, failure.Message("QRコードの画像に誤りがあります"))
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceDone(reserveID)
	if !ok {
		return failure.New(ErrScenario, failure.Message("QRコードのURLに誤りがあります"))
	}

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}

func SetShipment(ss *server.ServerShipment) {
	sShipment = ss
}

func SetPayment(sp *server.ServerPayment) {
	sPayment = sp
}
