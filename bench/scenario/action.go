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

	itemFromBuyerTrx, err := s2.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err := s1.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyer, err := s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err := s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != "trading" || itemFromSeller.Status != "trading" ||
		itemFromBuyerTrx.Status != "trading" || itemFromSellerTrx.Status != "trading" {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromSeller.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromBuyerTrx.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromSellerTrx.TransactionEvidenceStatus != "wait_shipping" {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != "initial" || itemFromSeller.ShippingStatus != "initial" ||
		itemFromBuyerTrx.ShippingStatus != "initial" || itemFromSellerTrx.ShippingStatus != "initial" {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	reserveID, apath, err := s1.Ship(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = s2.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err = s1.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != "trading" || itemFromSeller.Status != "trading" ||
		itemFromBuyerTrx.Status != "trading" || itemFromSellerTrx.Status != "trading" {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromSeller.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromBuyerTrx.TransactionEvidenceStatus != "wait_shipping" ||
		itemFromSellerTrx.TransactionEvidenceStatus != "wait_shipping" {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != "wait_pickup" || itemFromSeller.ShippingStatus != "wait_pickup" ||
		itemFromBuyerTrx.ShippingStatus != "wait_pickup" || itemFromSellerTrx.ShippingStatus != "wait_pickup" {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
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

	itemFromSellerTrx, err = s1.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = s2.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != "trading" || itemFromSeller.Status != "trading" ||
		itemFromBuyerTrx.Status != "trading" || itemFromSellerTrx.Status != "trading" {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != "wait_done" ||
		itemFromSeller.TransactionEvidenceStatus != "wait_done" ||
		itemFromBuyerTrx.TransactionEvidenceStatus != "wait_done" ||
		itemFromSellerTrx.TransactionEvidenceStatus != "wait_done" {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != "shipping" || itemFromSeller.ShippingStatus != "shipping" ||
		itemFromBuyerTrx.ShippingStatus != "shipping" || itemFromSellerTrx.ShippingStatus != "shipping" {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Message("配送予約IDに誤りがあります"))
	}

	err = s2.Complete(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err = s1.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = s2.FindItemFromUsesTransactions(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != "sold_out" || itemFromSeller.Status != "sold_out" ||
		itemFromBuyerTrx.Status != "sold_out" || itemFromSellerTrx.Status != "sold_out" {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != "done" ||
		itemFromSeller.TransactionEvidenceStatus != "done" ||
		itemFromBuyerTrx.TransactionEvidenceStatus != "done" ||
		itemFromSellerTrx.TransactionEvidenceStatus != "done" {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != "done" || itemFromSeller.ShippingStatus != "done" ||
		itemFromBuyerTrx.ShippingStatus != "done" || itemFromSellerTrx.ShippingStatus != "done" {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	return nil
}

func SetShipment(ss *server.ServerShipment) {
	sShipment = ss
}

func SetPayment(sp *server.ServerPayment) {
	sPayment = sp
}
