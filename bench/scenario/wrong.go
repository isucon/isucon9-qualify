package scenario

import (
	crand "crypto/rand"
	"fmt"

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

	csrfToken := s1.CSRFToken
	s1.CSRFToken = secureRandomStr(20)

	err = s1.SellWithWrongCSRFToken("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	// CSRFTokenを元に戻す
	s1.CSRFToken = csrfToken

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

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}
