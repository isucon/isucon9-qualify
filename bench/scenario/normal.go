package scenario

import (
	"fmt"
	"math/rand"

	"github.com/isucon/isucon9-qualify/bench/action"
	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	CorrectCardNumber = "AAAAAAAA"
	FailedCardNumber  = "FA10AAAA"
	IsucariShopID     = "11"

	ErrScenario failure.StringCode = "error scenario"
)

func initialize(paymentServiceURL, shipmentServiceURL string) (bool, error) {
	s1, err := session.NewSession()
	if err != nil {
		return false, err
	}

	return s1.Initialize(paymentServiceURL, shipmentServiceURL)
}

func sellAndBuy(user1, user2 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	s2, err := action.LoginedSession(user2)
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}
	token, err := s2.PaymentCard(CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}
	_, err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(ErrScenario, failure.Message("QRコードの画像に誤りがあります"))
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceDone(reserveID)
	if !ok {
		return failure.New(ErrScenario, failure.Message("QRコードのURLに誤りがあります"))
	}

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}

func loadSellNewCategoryBuyWithLoginedSession(s1, s2 *session.Session) error {
	targetItemID, err := s1.Sell("abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	err = newCategoryItemsWithLoginedSession(s2)
	if err != nil {
		return err
	}

	token, err := s2.PaymentCard(CorrectCardNumber, IsucariShopID)
	if err != nil {
		return err
	}
	_, err = s2.Buy(targetItemID, token)
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(targetItemID)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(ErrScenario, failure.Message("QRコードの画像に誤りがあります"))
	}

	err = s1.ShipDone(targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceDone(reserveID)
	if !ok {
		return failure.New(ErrScenario, failure.Message("QRコードのURLに誤りがあります"))
	}

	err = s2.Complete(targetItemID)
	if err != nil {
		return err
	}

	return nil
}

func transactionEvidence(user1 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	_, items, err := s1.UsersTransactions()
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if item.TransactionEvidenceStatus != ate.Status {
			return failure.New(ErrScenario, failure.Message("/users/transactions.jsonのステータスに誤りがあります"))
		}
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	_, items, err = s1.UsersTransactionsWithItemIDAndCreatedAt(targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(ErrScenario, failure.Message("/users/transactions.jsonのitem_idとcreated_atが正しく動作していません"))
		}

		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if item.TransactionEvidenceStatus != ate.Status {
			return failure.New(ErrScenario, failure.Message("/users/transactions.jsonのステータスに誤りがあります"))
		}
	}

	return nil
}

func userItemsAndItem(user1, user2 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	_, user, items, err := s1.UserItems(user2.ID)
	if err != nil {
		return err
	}

	for _, item := range items {
		aItem, ok := asset.GetItem(user.ID, item.ID)
		if !ok {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/users/%d.jsonに存在しない商品が返ってきています", user2.ID)))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/users/%d.jsonの商品の名前が間違えています", user2.ID)))
		}
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	_, user, items, err = s1.UserItemsWithItemIDAndCreatedAt(user2.ID, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/users/%d.jsonのitem_idとcreated_atが正しく動作していません", user2.ID)))
		}

		aItem, ok := asset.GetItem(user.ID, item.ID)
		if !ok {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/users/%d.jsonに存在しない商品が返ってきています", user2.ID)))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/users/%d.jsonの商品の名前が間違えています", user2.ID)))
		}
	}

	targetItemID = asset.GetUserItemsFirst(user2.ID)
	item, err := s1.Item(targetItemID)
	if err != nil {
		return err
	}

	aItem, ok := asset.GetItem(user2.ID, targetItemID)
	if !ok {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/items/%d.jsonに存在しない商品が返ってきています", targetItemID)))
	}

	if !(item.Description == aItem.Description) {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/items/%d.jsonの商品説明が間違っています", targetItemID)))
	}

	return nil
}

func bumpAndNewItems(user1, user2 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	s2, err := action.LoginedSession(user2)
	if err != nil {
		return err
	}

	targetItemID := asset.GetUserItemsFirst(user1.ID)
	newCreatedAt, err := s1.Bump(targetItemID)
	if err != nil {
		return err
	}

	asset.SetItemCreatedAt(user1.ID, targetItemID, newCreatedAt)

	hasNext, items, err := s2.NewItems()
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(ErrScenario, failure.Message("/new_items.jsonのhas_nextがfalseです"))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items.jsonの商品数が違います: expected: %d; actual: %d", asset.ItemsPerPage, len(items))))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	found := false
	for _, item := range items {
		if createdAt < item.CreatedAt {
			return failure.New(ErrScenario, failure.Message("/new_items.jsonはcreated_at順である必要があります"))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(ErrScenario, failure.Message("/new_items.jsonは販売中か売り切れの商品しか出してはいけません"))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(ErrScenario, failure.Message("/new_items.jsonで返している商品の情報に誤りがあります"))
		}

		if targetItemID == item.ID {
			found = true
		}

		createdAt = item.CreatedAt
	}

	if !found {
		// Verifyでしかできない確認
		return failure.New(ErrScenario, failure.Message("/new_items.jsonにバンプした商品が表示されていません"))
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	hasNext, items, err = s2.NewItemsWithItemIDAndCreatedAt(targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	if hasNext && (len(items) != asset.ItemsPerPage) {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items.jsonの商品数が違います: expected: %d; actual: %d", asset.ItemsPerPage, len(items))))
	}

	createdAt = items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(ErrScenario, failure.Message("/new_items.jsonのitem_idとcreated_atが正しく動作していません"))
		}

		if createdAt < item.CreatedAt {
			return failure.New(ErrScenario, failure.Message("/new_items.jsonはcreated_at順である必要があります"))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(ErrScenario, failure.Message("/new_items.jsonは販売中か売り切れの商品しか出してはいけません"))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(ErrScenario, failure.Message("/new_items.jsonで返している商品の情報に誤りがあります"))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func itemEdit(user1 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	targetItemID := asset.GetUserItemsFirst(user1.ID)
	price := 110
	_, err = s1.ItemEdit(targetItemID, price)
	if err != nil {
		return err
	}

	asset.SetItemPrice(user1.ID, targetItemID, price)

	return nil
}

func newCategoryItems(user1 asset.AppUser) error {
	s1, err := action.LoginedSession(user1)
	if err != nil {
		return err
	}

	category := asset.GetRandomRootCategory()

	hasNext, rootCategoryName, items, err := s1.NewCategoryItems(category.ID)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのhas_nextがfalseです", category.ID)))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items))))
	}

	if rootCategoryName != category.CategoryName {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのカテゴリ名が間違えています", category.ID)))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	for _, item := range items {
		if createdAt < item.CreatedAt {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonはcreated_at順である必要があります", category.ID)))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID)))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID)))
		}

		createdAt = item.CreatedAt
	}

	targetItemID, targetItemCreatedAt := items[len(items)-1].ID, items[len(items)-1].CreatedAt

	hasNext, rootCategoryName, items, err = s1.NewCategoryItemsWithItemIDAndCreatedAt(category.ID, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのhas_nextがfalseです", category.ID)))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items))))
	}

	if rootCategoryName != category.CategoryName {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのカテゴリ名が間違えています", category.ID)))
	}

	// 簡易チェック
	createdAt = items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのitem_idとcreated_atが正しく動作していません", category.ID)))
		}

		if createdAt < item.CreatedAt {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonはcreated_at順である必要があります", category.ID)))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID)))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID)))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func newCategoryItemsWithLoginedSession(s1 *session.Session) error {
	uitems := asset.GetUserItems(s1.UserID)
	tIndex := 0
	if len(uitems) >= 2 {
		tIndex = len(uitems) - rand.Intn(len(uitems)/2) - 1
	}

	targetItem, ok := asset.GetItem(s1.UserID, uitems[tIndex])
	if !ok {
		return failure.New(ErrScenario, failure.Message("/settingsのユーザーIDが存在しないIDです"))
	}

	category := asset.GetRandomRootCategory()
	_, _, items, err := s1.NewCategoryItemsWithItemIDAndCreatedAt(category.ID, targetItem.ID, targetItem.CreatedAt)
	if err != nil {
		return err
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items))))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItem.ID && item.CreatedAt <= targetItem.CreatedAt) {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonのitem_idとcreated_atが正しく動作していません", category.ID)))
		}

		if createdAt < item.CreatedAt {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonはcreated_at順である必要があります", category.ID)))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID)))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(ErrScenario, failure.Message(fmt.Sprintf("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID)))
		}

		createdAt = item.CreatedAt
	}

	return nil
}
