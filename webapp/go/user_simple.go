package main

import (
	"github.com/jmoiron/sqlx"
)

type UserSimple struct {
	ID           int64  `json:"id"`
	AccountName  string `json:"account_name"`
	NumSellItems int    `json:"num_sell_items"`
}

var users map[int64]UserSimple

// 揮発性のオンメモリキャッシュ
func getUserSimpleByID(q sqlx.Queryer, userID int64) (userSimple UserSimple, err error) {
	user, ok := users[userID]
	if ok {
		return user, nil
	}

	dbUser, err := getUserSimpleByID_internal(q, userID)
	return dbUser, err
}

// DBに問い合わせる用
func getUserSimpleByID_internal(q sqlx.Queryer, userID int64) (userSimple UserSimple, err error) {
	user := User{}
	err = sqlx.Get(q, &user, "SELECT * FROM `users` WHERE `id` = ?", userID)
	if err != nil {
		return userSimple, err
	}
	userSimple.ID = user.ID
	userSimple.AccountName = user.AccountName
	userSimple.NumSellItems = user.NumSellItems
	return userSimple, err
}
