package scenario

import (
	"context"
	"math/rand"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (bool, error) {
	s1, err := session.NewSession()
	if err != nil {
		return false, err
	}

	return s1.Initialize(ctx, paymentServiceURL, shipmentServiceURL)
}

func sellAndBuy(ctx context.Context, user1, user2 asset.AppUser) error {
	s1, err := LoginedSession(ctx, user1)
	if err != nil {
		return err
	}

	s2, err := LoginedSession(ctx, user2)
	if err != nil {
		return err
	}

	targetItemID, err := s1.Sell(ctx, "abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	err = buyComplete(ctx, s1, s2, targetItemID, 100)
	if err != nil {
		return err
	}

	return nil
}

func loadSellNewCategoryBuyWithLoginedSession(ctx context.Context, s1, s2 *session.Session) error {
	targetItemID, err := s1.Sell(ctx, "abcd", 100, "description description", 32)
	if err != nil {
		return err
	}

	err = newCategoryItemsWithLoginedSession(ctx, s2)
	if err != nil {
		return err
	}

	err = buyComplete(ctx, s1, s2, targetItemID, 100)
	if err != nil {
		return err
	}

	return nil
}

func transactionEvidence(ctx context.Context, user1 asset.AppUser) error {
	s1, err := LoginedSession(ctx, user1)
	if err != nil {
		return err
	}

	_, items, err := s1.UsersTransactions(ctx)
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
			return failure.New(fails.ErrApplication, failure.Message("/users/transactions.jsonのステータスに誤りがあります"))
		}
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	_, items, err = s1.UsersTransactionsWithItemIDAndCreatedAt(ctx, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(fails.ErrApplication, failure.Message("/users/transactions.jsonのitem_idとcreated_atが正しく動作していません"))
		}

		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if item.TransactionEvidenceStatus != ate.Status {
			return failure.New(fails.ErrApplication, failure.Message("/users/transactions.jsonのステータスに誤りがあります"))
		}
	}

	return nil
}

func userItemsAndItemWithLoginedSession(ctx context.Context, s1 *session.Session, userID int64) error {
	_, user, items, err := s1.UserItems(ctx, userID)
	if err != nil {
		return err
	}

	for _, item := range items {
		aItem, ok := asset.GetItem(user.ID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonに存在しない商品が返ってきています", userID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonの商品の名前が間違えています", userID))
		}
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	_, user, items, err = s1.UserItemsWithItemIDAndCreatedAt(ctx, userID, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonのitem_idとcreated_atが正しく動作していません", userID))
		}

		aItem, ok := asset.GetItem(user.ID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonに存在しない商品が返ってきています", userID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonの商品の名前が間違えています", userID))
		}
	}

	targetItemID = asset.GetUserItemsFirst(userID)
	item, err := s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	aItem, ok := asset.GetItem(userID, targetItemID)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonに存在しない商品が返ってきています", targetItemID))
	}

	if !(item.Description == aItem.Description) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	return nil
}

func bumpAndNewItemsWithLoginedSession(ctx context.Context, s1, s2 *session.Session) error {
	targetItemID := asset.GetUserItemsFirst(s1.UserID)
	newCreatedAt, err := s1.Bump(ctx, targetItemID)
	if err != nil {
		return err
	}

	asset.SetItemCreatedAt(s1.UserID, targetItemID, newCreatedAt)

	hasNext, items, err := s2.NewItems(ctx)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonのhas_nextがfalseです"))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの商品数が違います: expected: %d; actual: %d", asset.ItemsPerPage, len(items)))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	found := false
	for _, item := range items {
		if createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonはcreated_at順である必要があります"))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonは販売中か売り切れの商品しか出してはいけません"))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonで返している商品の情報に誤りがあります"))
		}

		if targetItemID == item.ID {
			found = true
		}

		createdAt = item.CreatedAt
	}

	if !found {
		// Verifyでしかできない確認
		return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonにバンプした商品が表示されていません"))
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	hasNext, items, err = s2.NewItemsWithItemIDAndCreatedAt(ctx, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	if hasNext && (len(items) != asset.ItemsPerPage) {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの商品数が違います: expected: %d; actual: %d", asset.ItemsPerPage, len(items)))
	}

	createdAt = items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonのitem_idとcreated_atが正しく動作していません"))
		}

		if createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonはcreated_at順である必要があります"))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonは販売中か売り切れの商品しか出してはいけません"))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(fails.ErrApplication, failure.Message("/new_items.jsonで返している商品の情報に誤りがあります"))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func itemEdit(ctx context.Context, user1 asset.AppUser) error {
	s1, err := LoginedSession(ctx, user1)
	if err != nil {
		return err
	}

	targetItemID := asset.GetUserItemsFirst(user1.ID)
	price := 110
	_, err = s1.ItemEdit(ctx, targetItemID, price)
	if err != nil {
		return err
	}

	asset.SetItemPrice(user1.ID, targetItemID, price)

	return nil
}

func newCategoryItems(ctx context.Context, user1 asset.AppUser) error {
	s1, err := LoginedSession(ctx, user1)
	if err != nil {
		return err
	}

	category := asset.GetRandomRootCategory()

	hasNext, rootCategoryName, items, err := s1.NewCategoryItems(ctx, category.ID)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのhas_nextがfalseです", category.ID))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items)))
	}

	if rootCategoryName != category.CategoryName {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのカテゴリ名が間違えています", category.ID))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	for _, item := range items {
		if createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonはcreated_at順である必要があります", category.ID))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID))
		}

		createdAt = item.CreatedAt
	}

	targetItemID, targetItemCreatedAt := items[len(items)-1].ID, items[len(items)-1].CreatedAt

	hasNext, rootCategoryName, items, err = s1.NewCategoryItemsWithItemIDAndCreatedAt(ctx, category.ID, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのhas_nextがfalseです", category.ID))
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items)))
	}

	if rootCategoryName != category.CategoryName {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのカテゴリ名が間違えています", category.ID))
	}

	// 簡易チェック
	createdAt = items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのitem_idとcreated_atが正しく動作していません", category.ID))
		}

		if createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonはcreated_at順である必要があります", category.ID))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func newCategoryItemsWithLoginedSession(ctx context.Context, s1 *session.Session) error {
	uitems := asset.GetUserItems(s1.UserID)
	tIndex := 0
	if len(uitems) >= 2 {
		tIndex = len(uitems) - rand.Intn(len(uitems)/2) - 1
	}

	targetItem, ok := asset.GetItem(s1.UserID, uitems[tIndex])
	if !ok {
		return failure.New(fails.ErrApplication, failure.Message("/settingsのユーザーIDが存在しないIDです"))
	}

	category := asset.GetRandomRootCategory()
	_, _, items, err := s1.NewCategoryItemsWithItemIDAndCreatedAt(ctx, category.ID, targetItem.ID, targetItem.CreatedAt)
	if err != nil {
		return err
	}

	if len(items) != asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの商品数が違います: expected: %d; actual: %d", category.ID, asset.ItemsPerPage, len(items)))
	}

	// 簡易チェック
	createdAt := items[0].CreatedAt
	for _, item := range items {
		if !(item.ID < targetItem.ID && item.CreatedAt <= targetItem.CreatedAt) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonのitem_idとcreated_atが正しく動作していません", category.ID))
		}

		if createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonはcreated_at順である必要があります", category.ID))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません", category.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name && aItem.Price == item.Price && aItem.Status == item.Status) {
			// TODO: aItem.CreatedAt == item.CreatedAtはinitializeを実装しないと確認できない
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります", category.ID))
		}

		createdAt = item.CreatedAt
	}

	return nil
}
