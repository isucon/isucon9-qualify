package scenario

import (
	"github.com/isucon/isucon9-qualify/bench/asset"
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
