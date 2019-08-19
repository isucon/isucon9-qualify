package scenario

import (
	"fmt"
	"net/http"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func irregularLoginWrongPassword(user1 asset.AppUser) error {
	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	err = s1.LoginWithWrongPassword(user1.AccountName, user1.Password+"wrong")
	if err != nil {
		return err
	}

	return nil
}

func irregularSellAndBuy(user1, user2, user3 asset.AppUser) error {
	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	seller, err := s1.Login(user1.AccountName, user1.Password)
	if err != nil {
		return err
	}

	if !user1.Equal(seller) {
		return failure.New(ErrScenario, failure.Message("ログインが失敗しています"))
	}

	err = s1.SetSettings()
	if err != nil {
		return err
	}

	err = s1.SellWithWrongCSRFToken("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	// 変な値段で買えない
	err = s1.SellWithWrongPrice("abcd", session.ItemMinPrice-1, "description description", 32)
	if err != nil {
		return err
	}

	err = s1.SellWithWrongPrice("abcd", session.ItemMaxPrice+1, "description description", 32)
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	s2, err := session.NewSession()
	if err != nil {
		return err
	}

	buyer, err := s2.Login(user2.AccountName, user2.Password)
	if err != nil {
		return err
	}

	if !user2.Equal(buyer) {
		return failure.New(ErrScenario, failure.Message("ログインが失敗しています"))
	}

	err = s2.SetSettings()
	if err != nil {
		return err
	}

	err = s1.BuyWithFailed(targetItemID, "", http.StatusForbidden, "自分の商品は買えません")
	if err != nil {
		return err
	}

	failedToken, err := s2.PaymentCard(FailedCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	err = s2.BuyWithFailed(targetItemID, failedToken, http.StatusBadRequest, "カードの残高が足りません")
	if err != nil {
		return err
	}

	token, err := s2.PaymentCard(CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	err = s2.BuyWithWrongCSRFToken(targetItemID, token)
	if err != nil {
		return err
	}

	s3, err := session.NewSession()
	if err != nil {
		return err
	}

	other, err := s3.Login(user3.AccountName, user3.Password)
	if err != nil {
		return err
	}

	if !user3.Equal(other) {
		return failure.New(ErrScenario, failure.Message("ログインが失敗しています"))
	}

	err = s3.SetSettings()
	if err != nil {
		return err
	}

	transactionEvidenceID, err := s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	oToken, err := s3.PaymentCard(CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	// onsaleでない商品は買えない
	err = s3.BuyWithFailed(targetItemID, oToken, http.StatusForbidden, "item is not for sale")
	if err != nil {
		return err
	}

	// QRコードはShipしないと見れない
	err = s1.DecodeQRURLWithFailed(fmt.Sprintf("/transactions/%d.png", transactionEvidenceID), http.StatusForbidden)
	if err != nil {
		return err
	}

	err = s1.ShipWithWrongCSRFToken(targetItemID)
	if err != nil {
		return err
	}

	// 他人はShipできない
	err = s3.ShipWithFailed(targetItemID, http.StatusForbidden, "権限がありません")
	if err != nil {
		return err
	}

	apath, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	// QRコードは他人だと見れない
	err = s3.DecodeQRURLWithFailed(apath, http.StatusForbidden)
	if err != nil {
		return err
	}

	surl, err := s1.DecodeQRURL(apath)
	if err != nil {
		return err
	}

	// acceptしない前はship_doneできない
	err = s1.ShipDoneWithFailed(targetItemID, http.StatusForbidden, "shipment service側で配送中か配送完了になっていません")
	if err != nil {
		return err
	}

	err = s3.ShipmentAccept(surl)
	if err != nil {
		return err
	}

	// 他人はship_doneできない
	err = s3.ShipDoneWithFailed(targetItemID, http.StatusForbidden, "権限がありません")
	if err != nil {
		return err
	}

	err = s1.ShipDoneWithWrongCSRFToken(targetItemID)
	if err != nil {
		return err
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	go func() {
	}()

	<-time.After(6 * time.Second)

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}
