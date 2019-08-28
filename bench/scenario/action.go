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
)

func activeSellerSession(ctx context.Context) (*session.Session, error) {
	s := ActiveSellerPool.Dequeue()
	if s != nil {
		return s, nil
	}

	user1 := asset.GetRandomActiveSeller()
	return loginedSession(ctx, user1)
}

func buyerSession(ctx context.Context) (*session.Session, error) {
	s := BuyerPool.Dequeue()
	if s != nil {
		return s, nil
	}

	user1 := asset.GetRandomBuyer()
	return loginedSession(ctx, user1)
}

func loginedSession(ctx context.Context, user1 asset.AppUser) (*session.Session, error) {
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

func sell(ctx context.Context, s1 *session.Session, price int) (int64, error) {
	name, description, categoryID := asset.GenText(8, false), asset.GenText(200, true), 32

	targetItemID, err := s1.Sell(ctx, name, price, description, categoryID)
	if err != nil {
		return 0, err
	}

	asset.SetItem(s1.UserID, targetItemID, name, price, description, categoryID)

	return targetItemID, nil
}

func buyComplete(ctx context.Context, s1, s2 *session.Session, targetItemID int64, price int) error {
	token := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

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
