package scenario

import (
	"context"
	"fmt"
	"net/http"
	"os"

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

func irregularSellAndBuy(ctx context.Context, s1, s2 *session.Session, user3 asset.AppUser) error {
	fileName, name, description, categoryID := asset.GetRandomImageFileName(), asset.GenText(8, false), asset.GenText(200, true), asset.GetRandomChildCategory().ID

	price := priceStoreCache.Get()

	err := s1.SellWithWrongCSRFToken(ctx, fileName, name, price, description, categoryID)
	if err != nil {
		return err
	}

	// 変な値段で買えない
	err = s1.SellWithWrongPrice(ctx, fileName, name, session.ItemMinPrice-1, description, categoryID)
	if err != nil {
		return err
	}

	err = s1.SellWithWrongPrice(ctx, fileName, name, session.ItemMaxPrice+1, description, categoryID)
	if err != nil {
		return err
	}

	targetParentCategoryID := asset.GetUser(s2.UserID).BuyParentCategoryID
	targetItemID, fileName, err := sellForFileName(ctx, s1, price, targetParentCategoryID)
	if err != nil {
		return err
	}

	f, err := os.Open(fileName)
	if err != nil {
		return failure.Wrap(err, failure.Message("ベンチマーカー内部のファイルを開くことに失敗しました"))
	}

	expectedMD5Str, err := calcMD5(f)
	if err != nil {
		return err
	}

	item, err := s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemMD5Str, err := s1.DownloadItemImageURL(ctx, item.ImageURL)
	if err != nil {
		return err
	}

	if expectedMD5Str != itemMD5Str {
		return failure.New(fails.ErrApplication, failure.Messagef("%sの画像のmd5値が間違っています expected: %s; actual: %s", item.ImageURL, expectedMD5Str, itemMD5Str))
	}

	err = s1.BuyWithFailed(ctx, targetItemID, "", http.StatusForbidden, "自分の商品は買えません")
	if err != nil {
		return err
	}

	failedToken := sPayment.ForceSet(FailedCardNumber, targetItemID, price)

	err = s2.BuyWithFailed(ctx, targetItemID, failedToken, http.StatusBadRequest, "カードの残高が足りません")
	if err != nil {
		return err
	}

	token := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

	err = s2.BuyWithWrongCSRFToken(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	s3, err := loginedSession(ctx, user3)
	if err != nil {
		return err
	}

	transactionEvidenceID, err := s2.Buy(ctx, targetItemID, token)
	if err != nil {
		return err
	}
	asset.UserBuyItem(s2.UserID)
	oToken := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

	// onsaleでない商品は買えない
	err = s3.BuyWithFailed(ctx, targetItemID, oToken, http.StatusForbidden, "item is not for sale")
	if err != nil {
		return err
	}

	// onsaleでない商品は編集できない
	err = s1.ItemEditWithNotOnSale(ctx, targetItemID, price+10)
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

	err = shipDone(ctx, s1, targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Message("集荷予約IDに誤りがあります"))
	}

	err = complete(ctx, s2, targetItemID)
	if err != nil {
		return err
	}

	return nil
}
