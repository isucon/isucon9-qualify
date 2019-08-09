package scenario

import (
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

func irregularSell(user1 asset.AppUser) error {
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

	return nil
}

func irregularBuy(user1, user2 asset.AppUser) error {
	s1, err := session.NewSession()
	if err != nil {
		return err
	}

	s2, err := session.NewSession()
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

	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}
	token, err := s2.PaymentCard(FailedCardNumber, IsucariShopID)
	if err != nil {
		return err
	}

	err = s2.BuyWithWrongCSRFToken(targetItemID, token)
	if err != nil {
		return err
	}

	err = s2.BuyWithFailedToken(targetItemID, token)
	if err != nil {
		return err
	}

	return nil
}
