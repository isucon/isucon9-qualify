package scenario

import (
	"context"
	"fmt"
	"net/http"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func irregularLoginWrongPassword(ctx context.Context, user1 asset.AppUser) error {
	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	err = s1.LoginWithWrongPassword(ctx, user1.AccountName, user1.Password+"wrong")
	if err != nil {
		return err
	}

	return nil
}

func irregularSellAndBuy(ctx context.Context, user1, user2, user3 asset.AppUser) error {
	s1, err := LoginedSession(ctx, user1)
	if err != nil {
		return err
	}

	err = s1.SellWithWrongCSRFToken(ctx, "abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	// 変な値段で買えない
	err = s1.SellWithWrongPrice(ctx, "abcd", session.ItemMinPrice-1, "description description", 32)
	if err != nil {
		return err
	}

	err = s1.SellWithWrongPrice(ctx, "abcd", session.ItemMaxPrice+1, "description description", 32)
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell(ctx, "abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	s2, err := LoginedSession(ctx, user2)
	if err != nil {
		return err
	}

	err = s1.BuyWithFailed(ctx, targetItemID, "", http.StatusForbidden, "自分の商品は買えません")
	if err != nil {
		return err
	}

	failedToken, err := s2.PaymentCard(ctx, FailedCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	err = s2.BuyWithFailed(ctx, targetItemID, failedToken, http.StatusBadRequest, "カードの残高が足りません")
	if err != nil {
		return err
	}

	token, err := s2.PaymentCard(ctx, CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	err = s2.BuyWithWrongCSRFToken(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	s3, err := LoginedSession(ctx, user3)
	if err != nil {
		return err
	}

	transactionEvidenceID, err := s2.Buy(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	oToken, err := s3.PaymentCard(ctx, CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	// onsaleでない商品は買えない
	err = s3.BuyWithFailed(ctx, targetItemID, oToken, http.StatusForbidden, "item is not for sale")
	if err != nil {
		return err
	}

	// QRコードはShipしないと見れない
	err = s1.DecodeQRURLWithFailed(ctx, fmt.Sprintf("/transactions/%d.png", transactionEvidenceID), http.StatusForbidden)
	if err != nil {
		return err
	}

	err = s1.ShipWithWrongCSRFToken(ctx, targetItemID)
	if err != nil {
		return err
	}

	// 他人はShipできない
	err = s3.ShipWithFailed(ctx, targetItemID, http.StatusForbidden, "権限がありません")
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(ctx, targetItemID)
	if err != nil {
		return err
	}

	// QRコードは他人だと見れない
	err = s3.DecodeQRURLWithFailed(ctx, apath, http.StatusForbidden)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(ctx, apath)
	if err != nil {
		return err
	}

	// acceptしない前はship_doneできない
	err = s1.ShipDoneWithFailed(ctx, targetItemID, http.StatusForbidden, "shipment service側で配送中か配送完了になっていません")
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(fails.ErrApplication, failure.Message("QRコードの画像に誤りがあります"))
	}

	// 他人はship_doneできない
	err = s3.ShipDoneWithFailed(ctx, targetItemID, http.StatusForbidden, "権限がありません")
	if err != nil {
		return err
	}

	err = s1.ShipDoneWithWrongCSRFToken(ctx, targetItemID)
	if err != nil {
		return err
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
