package scenario

import (
	"context"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	CorrectCardNumber = "AAAAAAAA"
	FailedCardNumber  = "FA10AAAA"
	IsucariShopID     = "11"
)

func LoginedSession(ctx context.Context, user1 asset.AppUser) (*session.Session, error) {
	s1, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	user, err := s1.Login(ctx, user1.AccountName, user1.Password)
	if err != nil {
		return nil, err
	}

	if !user1.Equal(user) {
		return nil, failure.New(fails.ErrApplication, failure.Message("ログインが失敗しています"))
	}

	err = s1.SetSettings(ctx)
	if err != nil {
		return nil, err
	}

	return s1, nil
}

func buyComplete(ctx context.Context, s1, s2 *session.Session, targetItemID int64) error {
	token := sPayment.ForceSet(CorrectCardNumber)

	_, err := s2.Buy(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(ctx, targetItemID)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(ctx, apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(fails.ErrApplication, failure.Message("QRコードの画像に誤りがあります"))
	}

	err = s1.ShipDone(ctx, targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Message("配送予約IDに誤りがあります"))
	}

	err = s2.Complete(ctx, targetItemID)
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
