package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (bool, error) {
	s1, err := session.NewSessionForInialize()
	if err != nil {
		return false, err
	}

	return s1.Initialize(ctx, paymentServiceURL, shipmentServiceURL)
}

func checkItemSimpleCategory(item session.ItemSimple, aItem asset.AppItem) error {
	aCategory, _ := asset.GetCategory(aItem.CategoryID)
	aRootCategory, _ := asset.GetCategory(aCategory.ParentID)

	if item.Category.ID == 0 {
		return fmt.Errorf("商品のカテゴリーIDがありません")
	}
	if item.Category.ID != aCategory.ID || item.Category.CategoryName != aCategory.CategoryName {
		return fmt.Errorf("商品のカテゴリーが異なります")
	}
	if item.Category.ParentID == 0 {
		return fmt.Errorf("商品の親カテゴリーIDがありません")
	}
	if item.Category.ParentID != aRootCategory.ID || item.Category.ParentCategoryName != aRootCategory.CategoryName {
		return fmt.Errorf("商品の親カテゴリーが異なります")
	}

	return nil
}

func checkItemDetailCategory(item session.ItemDetail, aItem asset.AppItem) error {
	aCategory, _ := asset.GetCategory(aItem.CategoryID)
	aRootCategory, _ := asset.GetCategory(aCategory.ParentID)

	if item.Category.ID == 0 {
		return fmt.Errorf("商品のカテゴリーIDがありません")
	}
	if item.Category.ID != aCategory.ID || item.Category.CategoryName != aCategory.CategoryName {
		return fmt.Errorf("商品のカテゴリーが異なります")
	}
	if item.Category.ParentID == 0 {
		return fmt.Errorf("商品の親カテゴリーIDがありません")
	}
	if item.Category.ParentID != aRootCategory.ID || item.Category.ParentCategoryName != aRootCategory.CategoryName {
		return fmt.Errorf("商品の親カテゴリーが異なります")
	}

	return nil
}

func loadSellNewCategoryBuyWithLoginedSession(ctx context.Context, s1, s2 *session.Session, price int) error {
	targetItemID, err := sell(ctx, s1, price)
	if err != nil {
		return err
	}

	err = loadNewCategoryItemsWithLoginedSession(ctx, s1)
	if err != nil {
		return err
	}

	err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
	if err != nil {
		return err
	}

	return nil
}

func transactionEvidence(ctx context.Context, s1 *session.Session) error {
	hasNext, items, err := s1.UsersTransactions(ctx)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのhas_nextがfalseになっています (user_id: %d)", s1.UserID))
	}

	var createdAt int64
	for _, item := range items {
		if createdAt > 0 && createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Message("/users/transactions.jsonはcreated_at順である必要があります"))
		}
		createdAt = item.CreatedAt

		aItem, ok := asset.GetItem(item.SellerID, item.ID)

		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに存在しない商品が返ってきています (item_id: %d)", item.ID))
		}

		if !(item.Description == aItem.Description) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品説明が間違っています (item_id: %d)", item.ID))
		}

		err = checkItemDetailCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate, ok := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if ok && item.TransactionEvidenceStatus != ate.Status {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのステータスに誤りがあります (user_id: %d)", s1.UserID))
		}
	}

	targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

	_, items, err = s1.UsersTransactionsWithItemIDAndCreatedAt(ctx, targetItemID, targetItemCreatedAt)
	if err != nil {
		return err
	}

	for _, item := range items {
		if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのitem_idとcreated_atが正しく動作していません (user_id: %d)", s1.UserID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)

		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに存在しない商品が返ってきています (item_id: %d)", item.ID))
		}

		if !(item.Description == aItem.Description) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品説明が間違っています (item_id: %d)", item.ID))
		}

		err = checkItemDetailCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s (itme_id: %d)", err.Error(), item.ID))
		}

		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate, ok := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if ok && item.TransactionEvidenceStatus != ate.Status {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのステータスに誤りがあります (user_id: %d)", s1.UserID))
		}
	}

	return nil
}

func loadTransactionEvidence(ctx context.Context, s1 *session.Session) error {
	hasNext, items, err := s1.UsersTransactions(ctx)
	if err != nil {
		return err
	}

	if !hasNext {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのhas_nextがfalseになっています (user_id: %d)", s1.UserID))
	}

	for _, item := range items {
		aItem, ok := asset.GetItem(item.SellerID, item.ID)

		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに存在しない商品が返ってきています (item_id: %d)", item.ID))
		}

		if !(item.Description == aItem.Description) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品説明が間違っています (item_id: %d)", item.ID))
		}

		err = checkItemDetailCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		if item.TransactionEvidenceID == 0 {
			// TODO: check
			continue
		}

		ate, ok := asset.GetTransactionEvidence(item.TransactionEvidenceID)
		if ok && item.TransactionEvidenceStatus != ate.Status {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのステータスに誤りがあります (user_id: %d)", s1.UserID))
		}
	}

	for {
		targetItemID, targetItemCreatedAt := items[len(items)/2].ID, items[len(items)/2].CreatedAt

		hasNext, items, err = s1.UsersTransactionsWithItemIDAndCreatedAt(ctx, targetItemID, targetItemCreatedAt)
		if err != nil {
			return err
		}

		if !hasNext {
			// TODO: check
			break
		}

		for _, item := range items {
			if !(item.ID < targetItemID && item.CreatedAt <= targetItemCreatedAt) {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのitem_idとcreated_atが正しく動作していません (user_id: %d)", s1.UserID))
			}

			aItem, ok := asset.GetItem(item.SellerID, item.ID)

			if !ok {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに存在しない商品が返ってきています (item_id: %d)", item.ID))
			}

			if !(item.Description == aItem.Description) {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品説明が間違っています (item_id: %d)", item.ID))
			}

			err = checkItemDetailCategory(item, aItem)
			if err != nil {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s (item_id: %d)", err.Error(), item.ID))
			}

			if item.TransactionEvidenceID == 0 {
				// TODO: check
				continue
			}

			ate, ok := asset.GetTransactionEvidence(item.TransactionEvidenceID)
			if ok && item.TransactionEvidenceStatus != ate.Status {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのステータスに誤りがあります (user_id: %d)", s1.UserID))
			}
		}
	}

	return nil
}

func getItem(ctx context.Context, s *session.Session, targetItemID int64) error {
	item, err := s.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	aItem, ok := asset.GetItem(item.SellerID, item.ID)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonに存在しない商品が返ってきています", targetItemID))
	}

	if !(item.Description == aItem.Description) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	err = checkItemDetailCategory(*item, aItem)
	if err != nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの%s", targetItemID, err.Error()))
	}

	err = checkItemDetailCategory(*item, aItem)
	if err != nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s (item_id: %d)", err.Error(), item.ID))
	}

	return nil
}

// ユーザページの商品をすべて数える
// カテゴリやタイトルも確認しているので、active sellerをある程度の人数確認したら
// 商品が大幅に削除されていないことも確認できる
func userItemsAllAndItems(ctx context.Context, s *session.Session, sellerID int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := itemIDsFromUsers(ctx, s, itemIDs, sellerID, 0, 0, 0)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	buffer := 10 // TODO
	aUser := asset.GetUser(sellerID)
	if aUser.NumSellItems > c+buffer || aUser.NumSellItems < c-buffer || c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の商品数が正しくありません", sellerID))
	}
	if checkItem == 0 {
		return nil
	}
	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := getItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func itemIDsFromUsers(ctx context.Context, s *session.Session, itemIDs *IDsStore, sellerID, nextItemID, nextCreatedAt, loop int64) error {
	var hasNext bool
	var items []session.ItemSimple
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, _, items, err = s.UserItemsWithItemIDAndCreatedAt(ctx, sellerID, nextItemID, nextCreatedAt)
	} else {
		hasNext, _, items, err = s.UserItems(ctx, sellerID)
	}
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.SellerID != sellerID {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の出品者が正しくありません　(item_id: %d)", sellerID, item.ID))
		}

		aItem, ok := asset.GetItem(sellerID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonに存在しない商品 (item_id: %d) が返ってきています", sellerID, item.ID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonの商品の名前が間違えています", sellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonの%s", sellerID, err.Error()))
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonに同じ商品がありました (item_id: %d)", sellerID, item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if hasNext && loop < 100 { // TODO: max pager
		err := itemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return err
		}
	}
	return nil
}

func findItemFromUsersTransactions(ctx context.Context, s *session.Session, targetItemID int64) (session.ItemDetail, error) {
	return findItemFromUsersTransactionsAll(ctx, s, targetItemID, 0, 0, 0)
}

func findItemFromUsersTransactionsAll(ctx context.Context, s *session.Session, targetItemID, nextItemID, nextCreatedAt, loop int64) (session.ItemDetail, error) {
	var hasNext bool
	var items []session.ItemDetail
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, items, err = s.UsersTransactionsWithItemIDAndCreatedAt(ctx, nextItemID, nextCreatedAt)
	} else {
		hasNext, items, err = s.UsersTransactions(ctx)
	}
	if err != nil {
		return session.ItemDetail{}, err
	}

	for _, item := range items {
		if item.ID == targetItemID {
			return item, nil
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if hasNext && loop < 100 { // TODO: max pager
		_, err := findItemFromUsersTransactionsAll(ctx, s, targetItemID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return session.ItemDetail{}, err
		}
	}
	return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json から商品を探すことができませんでした　(item_id: %d)", targetItemID))
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
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonは販売中か売り切れの商品しか出してはいけません (item_id: %d; seller_id: %d)", item.ID, item.SellerID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの商品情報に誤りがあります (item_id: %d; seller_id: %d)", item.ID, item.SellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		if targetItemID == item.ID {
			found = true
		}

		createdAt = item.CreatedAt
	}

	if !found {
		// Verifyでしかできない確認
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonにバンプした商品が表示されていません (item_id: %d)", targetItemID))
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
		if ok && !(aItem.Name == item.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの商品情報に誤りがあります (item_id: %d; seller_id: %d)", item.ID, item.SellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func itemEditWithLoginedSession(ctx context.Context, s1 *session.Session, targetItemID int64, price int) error {
	_, err := s1.ItemEdit(ctx, targetItemID, price)
	if err != nil {
		return err
	}

	asset.SetItemPrice(s1.UserID, targetItemID, price)

	return nil
}

func itemEditNewItemWithLoginedSession(ctx context.Context, s1 *session.Session, targetItemID int64, price int) error {
	_, err := s1.ItemEdit(ctx, targetItemID, price)
	if err != nil {
		return err
	}

	return nil
}

func newCategoryItemsWithLoginedSession(ctx context.Context, s1 *session.Session) error {
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
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		aCategory, _ := asset.GetCategory(aItem.CategoryID)

		if ok && !(aItem.Name == item.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		if category.ID != aCategory.ParentID {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品のカテゴリに誤りがあります (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの%s (item_id: %d)", category.ID, err.Error(), item.ID))
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
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

func loadNewCategoryItemsWithLoginedSession(ctx context.Context, s1 *session.Session) error {
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
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonは販売中か売り切れの商品しか出してはいけません (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if ok && !(aItem.Name == item.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonで返している商品の情報に誤りがあります (item_id: %d; seller_id: %d)", category.ID, item.ID, item.SellerID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonの%s (item_id: %d)", err.Error(), item.ID))
		}

		createdAt = item.CreatedAt
	}

	return nil
}

type IDsStore struct {
	sync.RWMutex
	ids map[int64]bool
}

func newIDsStore() *IDsStore {
	m := make(map[int64]bool)
	s := &IDsStore{
		ids: m,
	}
	return s
}

func (s *IDsStore) Add(id int64) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.ids[id]; ok {
		return fmt.Errorf("duplicated ID found: %d", id)
	}
	s.ids[id] = true
	return nil
}

func (s *IDsStore) Len() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.ids)
}

func (s *IDsStore) RandomIDs(num int) []int64 {
	s.RLock()
	defer s.RUnlock()
	if len(s.ids) < num {
		num = len(s.ids)
	}
	ids := make([]int64, 0, len(s.ids))
	for id := range s.ids {
		ids = append(ids, id)
	}
	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	return ids[0:num]
}
