package scenario

import (
	"net/http"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
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
		return fails.NewError(nil, "ログインが失敗しています")
	}

	err = s1.SetSettings()
	if err != nil {
		return err
	}

	err = s1.SellWithWrongCSRFToken("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

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
		return fails.NewError(nil, "ログインが失敗しています")
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
		return fails.NewError(nil, "ログインが失敗しています")
	}

	err = s3.SetSettings()
	if err != nil {
		return err
	}

	err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	err = s3.BuyWithFailed(targetItemID, token, http.StatusForbidden, "item is not for sale")
	if err != nil {
		return err
	}

	err = s1.ShipWithWrongCSRFToken(targetItemID)
	if err != nil {
		return err
	}

	return nil
}
