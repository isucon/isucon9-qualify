package scenario

import (
	"context"
	"os"
	"sync"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func Verify(ctx context.Context) *fails.Critical {
	var wg sync.WaitGroup

	critical := fails.NewCritical()

	// verify scenario #1
	wg.Add(1)
	go func() {
		defer wg.Done()

		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		targetItemID, fileName, err := sellForFileName(ctx, s1, 100)
		if err != nil {
			critical.Add(err)
			return
		}

		f, err := os.Open(fileName)
		if err != nil {
			critical.Add(failure.Wrap(err, failure.Message("ベンチマーカー内部のファイルを開くことに失敗しました")))
			return
		}

		expectedMD5Str, err := calcMD5(f)
		if err != nil {
			critical.Add(err)
			return
		}

		item, err := s1.Item(ctx, targetItemID)
		if err != nil {
			critical.Add(err)
			return
		}

		md5Str, err := s1.DownloadItemImageURL(ctx, item.ImageURL)
		if err != nil {
			critical.Add(err)
			return
		}

		if expectedMD5Str != md5Str {
			critical.Add(failure.New(fails.ErrApplication, failure.Messagef("%sの画像のmd5値が間違っています expected: %s; actual: %s", item.ImageURL, expectedMD5Str, md5Str)))
			return
		}

		err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, 100)
		if err != nil {
			critical.Add(err)
			return
		}
	}()

	// verify scenario #2
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		err = verifyBumpAndNewItems(ctx, s1, s2)
		if err != nil {
			critical.Add(err)
		}
	}()

	// verify scenario #3
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		category := asset.GetRandomRootCategory()
		err = verifyNewCategoryItemsAndItems(ctx, s2, category.ID, 2, 20)
		if err != nil {
			critical.Add(err)
		}

		targetItemID := asset.GetUserItemsFirst(s1.UserID)
		err = itemEditWithLoginedSession(ctx, s1, targetItemID, 110)
		if err != nil {
			critical.Add(err)
		}
	}()

	// verify scenario #4
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s1)
		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		err = verifyTransactionEvidence(ctx, s1, 3, 27)
		if err != nil {
			critical.Add(err)
			return
		}

		targetItem, err := sell(ctx, s1, 100)
		if err != nil {
			critical.Add(err)
			return
		}
		_, err = findItemFromUsers(ctx, s1, targetItem, 1)
		if err != nil {
			critical.Add(err)
			return
		}
		_, err = findItemFromNewCategory(ctx, s1, targetItem, 1)
		if err != nil {
			critical.Add(err)
			return
		}
		_, err = findItemFromUsers(ctx, s2, targetItem, 1)
		if err != nil {
			critical.Add(err)
			return
		}
		_, err = findItemFromNewCategory(ctx, s2, targetItem, 1)
		if err != nil {
			critical.Add(err)
			return
		}
		_, err = findItemFromUsersTransactions(ctx, s1, targetItem.ID, 1)
		if err != nil {
			critical.Add(err)
			return
		}

		err = verifyTransactionEvidence(ctx, s1, 2, 5)
		if err != nil {
			critical.Add(err)
			return
		}

		err = buyCompleteWithVerify(ctx, s1, s2, targetItem.ID, 100)
		if err != nil {
			critical.Add(err)
		}
	}()

	// verify scenario #5
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		s2, err := activeSellerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer ActiveSellerPool.Enqueue(s2)

		// buyer の全件確認 (self)
		err = verifyUserItemsAndItems(ctx, s1, s1.UserID, 0)
		if err != nil {
			critical.Add(err)
			return
		}

		// active sellerの全件確認(self)
		err = verifyUserItemsAndItems(ctx, s2, s2.UserID, 10)
		if err != nil {
			critical.Add(err)
			return
		}

		// active sellerではないユーザも確認。0件でも問題ない
		userIDs := asset.GetRandomBuyerIDs(3)
		for _, userID := range userIDs {
			err = verifyUserItemsAndItems(ctx, s1, userID, 0)
			if err != nil {
				critical.Add(err)
			}
		}
	}()

	// verify scenario #6
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		// active sellerの全件確認(random)
		userIDs := asset.GetRandomActiveSellerIDs(3)
		for _, userID := range userIDs {
			err = verifyUserItemsAndItems(ctx, s1, userID, 5)
			if err != nil {
				critical.Add(err)
			}
		}

	}()

	user3 := asset.GetRandomBuyer()

	// verify scenario #7
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := irregularLoginWrongPassword(ctx, user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	// verify scenario #8
	wg.Add(1)
	go func() {
		defer wg.Done()
		s1, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s1)

		s2, err := buyerSession(ctx)
		if err != nil {
			critical.Add(err)
			return
		}
		defer BuyerPool.Enqueue(s2)

		err = irregularSellAndBuy(ctx, s1, s2, user3)
		if err != nil {
			critical.Add(err)
		}
	}()

	wg.Wait()

	return critical
}

func verifyBumpAndNewItems(ctx context.Context, s1, s2 *session.Session) error {
	targetItemID := asset.GetUserItemsFirst(s1.UserID)
	newCreatedAt, err := s1.Bump(ctx, targetItemID)
	if err != nil {
		return err
	}

	targetItem := asset.SetItemCreatedAt(s1.UserID, targetItemID, newCreatedAt)

	itemFromNewCategory, err := findItemFromNewCategory(ctx, s1, targetItem, 1)
	if err != nil {
		return err
	}
	if itemFromNewCategory.CreatedAt != newCreatedAt {
		return failure.New(fails.ErrApplication, failure.Messagef("Bump後の商品が更新されていません (item_id: %d)", targetItemID))
	}
	itemFromUsers, err := findItemFromUsers(ctx, s1, targetItem, 1)
	if err != nil {
		return err
	}
	if itemFromUsers.CreatedAt != newCreatedAt {
		return failure.New(fails.ErrApplication, failure.Messagef("Bump後の商品が更新されていません (item_id: %d)", targetItemID))
	}

	targetCategory, ok := asset.GetCategory(targetItem.CategoryID)
	if !ok || targetCategory.ParentID == 0 {
		// データ不整合・ベンチマーカのバグの可能性
		return failure.New(fails.ErrApplication, failure.Messagef("商品のカテゴリを探すことができませんでした (item_id: %d)", targetItem.ID))
	}

	err = verifyNewCategoryItemsAndItems(ctx, s1, targetCategory.ParentID, 2, 5)
	if err != nil {
		return err
	}

	return nil
}

// カテゴリページの商品をたどる
func verifyNewCategoryItemsAndItems(ctx context.Context, s *session.Session, categoryID int, maxPage int64, checkItem int) error {
	category, ok := asset.GetCategory(categoryID)
	if !ok || category.ParentID != 0 {
		// benchmarkerのバグになるかと
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json カテゴリIDが正しくありません", categoryID))
	}
	itemIDs := newIDsStore()
	err := verifyItemIDsFromCategory(ctx, s, itemIDs, categoryID, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// 全件チェックの時だけチェック
	// countUserItemsでもチェックしている。商品数perpage*maxpageの98%あればよい
	if (maxPage == 0 && c < 3000) || float64(c) < float64(maxPage)*float64(asset.ItemsPerPage)*0.98 { // TODO 98%?
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json の商品数が正しくありません", categoryID))
	}

	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := verifyGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func verifyItemIDsFromCategory(ctx context.Context, s *session.Session, itemIDs *IDsStore, categoryID int, nextItemID, nextCreatedAt, loop, maxPage int64) error {
	var hasNext bool
	var items []session.ItemSimple
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, _, items, err = s.NewCategoryItemsWithItemIDAndCreatedAt(ctx, categoryID, nextItemID, nextCreatedAt)
	} else {
		hasNext, _, items, err = s.NewCategoryItems(ctx, categoryID)
	}
	if err != nil {
		return err
	}
	if loop < 50 && asset.ItemsPerPage != len(items) { // MEMO 50件よりはみないだろう
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json の商品数が正しくありません", categoryID))
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonはcreated_at順である必要があります", categoryID))
		}

		if item.Category.ParentID != categoryID {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json のカテゴリが異なります (item_id: %d)", categoryID, item.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonに不明な商品があります (item_id: %d)", categoryID, item.ID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonの商品の名前が間違えています (item_id: %d)", categoryID, item.ID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonの%s", categoryID, err.Error()))
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonに同じ商品がありました (item_id: %d)", categoryID, item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return nil
	}
	if hasNext && loop < 100 { // TODO: max pager
		return verifyItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
	}
	return nil
}

func verifyTransactionEvidence(ctx context.Context, s *session.Session, maxPage int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := verifyItemIDsTransactionEvidence(ctx, s, itemIDs, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// todo assetsからとれるか
	if c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品数が正しくありません (user_id: %d)", s.UserID))
	}
	if checkItem == 0 {
		return nil
	}
	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := verifyGetItemTE(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyItemIDsTransactionEvidence(ctx context.Context, s *session.Session, itemIDs *IDsStore, nextItemID, nextCreatedAt, loop, maxPage int64) error {
	var hasNext bool
	var items []session.ItemDetail
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, items, err = s.UsersTransactionsWithItemIDAndCreatedAt(ctx, nextItemID, nextCreatedAt)
	} else {
		hasNext, items, err = s.UsersTransactions(ctx)
	}
	if err != nil {
		return err
	}
	if hasNext && asset.ItemsTransactionsPerPage != len(items) {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品数が正しくありません (user_id: %d)", s.UserID))
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonはcreated_at順である必要があります"))
		}

		if item.BuyerID != s.UserID && item.Seller.ID != s.UserID {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに購入・出品していない商品が含まれます (item_id: %d, user_id: %d)", item.ID, s.UserID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに不明な商品があります (item_id: %d)", item.ID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品の名前が間違えています (item_id: %d)", item.ID))
		}

		err := checkItemDetailCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s", err.Error()))
		}

		if item.BuyerID == s.UserID && (item.SellerID == s.UserID && item.BuyerID != 0) {
			if item.TransactionEvidenceID == 0 {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのTransactionEvidence情報が正しくありません (item_id: %d, user_id: %d)", item.ID, s.UserID))
			}
			if item.TransactionEvidenceStatus == "" {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのTransactionEvidence情報が正しくありません (item_id: %d, user_id: %d)", item.ID, s.UserID))
			}
			if item.ShippingStatus == "" {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのshipping情報が正しくありません (item_id: %d, user_id: %d)", item.ID, s.UserID))
			}

			ate, ok := asset.GetTransactionEvidence(item.TransactionEvidenceID)
			if ok && item.TransactionEvidenceStatus != ate.Status {
				return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonのステータスに誤りがあります (item_id: %d, user_id: %d)", item.ID, s.UserID))
			}
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに同じ商品がありました (item_id: %d, user_id: %d)", item.ID, s.UserID))
		}

		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return nil
	}
	if hasNext && loop < 100 { // TODO: max pager
		return verifyItemIDsTransactionEvidence(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil
}

// ユーザページをたどる
func verifyUserItemsAndItems(ctx context.Context, s *session.Session, sellerID int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := verifyItemIDsFromUsers(ctx, s, itemIDs, sellerID, 0, 0, 0)
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
		err := verifyGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func verifyItemIDsFromUsers(ctx context.Context, s *session.Session, itemIDs *IDsStore, sellerID, nextItemID, nextCreatedAt, loop int64) error {
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
	// 件数チェックはしない。合計でチェックしている
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonはcreated_at順である必要があります", sellerID))
		}

		if item.SellerID != sellerID {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の出品者が正しくありません　(item_id: %d)", sellerID, item.ID))
		}

		aItem, ok := asset.GetItem(sellerID, item.ID)
		if !ok {
			continue
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
		return verifyItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
	}
	return nil
}

func verifyGetItem(ctx context.Context, s *session.Session, targetItemID int64) error {
	item, err := s.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	if !(item.Description != "") {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	if item.Seller.ID != item.SellerID {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの出品者情報が正しくありません", targetItemID))
	}

	aItem, ok := asset.GetItem(item.SellerID, item.ID)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonは不明な商品です", targetItemID))

	}

	if !(item.Name == aItem.Name) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品名が間違っています", targetItemID))
	}
	if !(item.Description == aItem.Description) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}
	if !(item.ImageURL == getImageURL(aItem.ImageName)) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品画像URLが間違っています", targetItemID))
	}

	md5Str, err := s.DownloadItemImageURL(ctx, item.ImageURL)
	if err != nil {
		return err
	}

	expectedMD5 := asset.GetImageMD5(aItem.ImageName)
	if expectedMD5 != md5Str {
		return failure.New(fails.ErrApplication, failure.Messagef("%sの商品画像が間違っています", item.ImageURL))
	}

	err = checkItemDetailCategory(item, aItem)
	if err != nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの%s", targetItemID, err.Error()))
	}

	if item.BuyerID != 0 && item.Buyer == nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonのbuyer_idがあるのに購入者の情報がありません", targetItemID))
	}

	if item.BuyerID != 0 && item.Buyer.ID != item.BuyerID {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの購入者情報が正しくありません", targetItemID))
	}

	return nil
}

func verifyGetItemTE(ctx context.Context, s *session.Session, targetItemID int64) error {
	item, err := s.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	if !(item.Description != "") {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	if item.Seller.ID != item.SellerID {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの出品者情報が正しくありません", targetItemID))
	}

	aItem, ok := asset.GetItem(item.SellerID, item.ID)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonは不明な商品です", targetItemID))

	}

	if !(item.Name == aItem.Name) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品名が間違っています", targetItemID))
	}
	if !(item.Description == aItem.Description) {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	err = checkItemDetailCategory(item, aItem)
	if err != nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの%s", targetItemID, err.Error()))
	}

	if item.BuyerID != 0 && item.Buyer == nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonのbuyer_idがあるのに購入者の情報がありません", targetItemID))
	}

	if item.BuyerID != 0 && item.Buyer.ID != item.BuyerID {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの購入者情報が正しくありません", targetItemID))
	}

	if item.BuyerID == s.UserID && (item.SellerID == s.UserID && item.BuyerID != 0) {
		if item.TransactionEvidenceID == 0 {
			return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonのTransactionEvidence情報が正しくありません", targetItemID))
		}
		if item.TransactionEvidenceStatus == "" {
			return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonのTransactionEvidence情報が正しくありません", targetItemID))
		}
		if item.ShippingStatus == "" {
			return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonのshipping情報が正しくありません", targetItemID))
		}
	}

	return nil
}
