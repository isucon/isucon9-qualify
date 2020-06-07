package main

import (
	"errors"
	"github.com/jmoiron/sqlx"
)

var categories map[int]Category = map[int]Category{
	1:  {1, 0, "ソファー", ""},
	2:  {2, 1, "一人掛けソファー", ""},
	3:  {3, 1, "二人掛けソファー", ""},
	4:  {4, 1, "コーナーソファー", ""},
	5:  {5, 1, "二段ソファー", ""},
	6:  {6, 1, "ソファーベッド", ""},
	10: {10, 0, "家庭用チェア", ""},
	11: {11, 10, "スツール", ""},
	12: {12, 10, "クッションスツール", ""},
	13: {13, 10, "ダイニングチェア", ""},
	14: {14, 10, "リビングチェア", ""},
	15: {15, 10, "カウンターチェア", ""},
	20: {20, 0, "キッズチェア", ""},
	21: {21, 20, "学習チェア", ""},
	22: {22, 20, "ベビーソファ", ""},
	23: {23, 20, "キッズハイチェア", ""},
	24: {24, 20, "テーブルチェア", ""},
	30: {30, 0, "オフィスチェア", ""},
	31: {31, 30, "デスクチェア", ""},
	32: {32, 30, "ビジネスチェア", ""},
	33: {33, 30, "回転チェア", ""},
	34: {34, 30, "リクライニングチェア", ""},
	35: {35, 30, "投擲用椅子", ""},
	40: {40, 0, "折りたたみ椅子", ""},
	41: {41, 40, "パイプ椅子", ""},
	42: {42, 40, "木製折りたたみ椅子", ""},
	43: {43, 40, "キッチンチェア", ""},
	44: {44, 40, "アウトドアチェア", ""},
	45: {45, 40, "作業椅子", ""},
	50: {50, 0, "ベンチ", ""},
	51: {51, 50, "一人掛けベンチ", ""},
	52: {52, 50, "二人掛けベンチ", ""},
	53: {53, 50, "アウトドア用ベンチ", ""},
	54: {54, 50, "収納付きベンチ", ""},
	55: {55, 50, "背もたれ付きベンチ", ""},
	56: {56, 50, "ベンチマーク", ""},
	60: {60, 0, "座椅子", ""},
	61: {61, 60, "和風座椅子", ""},
	62: {62, 60, "高座椅子", ""},
	63: {63, 60, "ゲーミング座椅子", ""},
	64: {64, 60, "ロッキングチェア", ""},
	65: {65, 60, "座布団", ""},
}

type Category struct {
	ID                 int    `json:"id" db:"id"`
	ParentID           int    `json:"parent_id" db:"parent_id"`
	CategoryName       string `json:"category_name" db:"category_name"`
	ParentCategoryName string `json:"parent_category_name,omitempty" db:"-"`
}

// エラーメッセージは適当です
// もしかしたらsqlxのエラーメッセージと同じものを返す必要あり?
func getCategoryByID(q sqlx.Queryer, categoryID int) (category Category, err error) {
	targetCategory, ok1 := categories[categoryID]
	if !ok1 {
		return Category{}, errors.New("not found")
	}

	if targetCategory.ParentID != 0 {
		parentCategory, ok2 := categories[targetCategory.ParentID]
		if !ok2 {
			return Category{}, errors.New("not found")
		}
		targetCategory.ParentCategoryName = parentCategory.CategoryName
	}

	return targetCategory, nil
}
