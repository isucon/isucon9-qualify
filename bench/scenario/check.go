package scenario

import (
	"context"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

func Check(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	user3 := asset.GetRandomBuyer()

	// check scenario #1
	// 間違ったパスワードでログインができないことをチェックする
	// これがないとパスワードチェックを外して常にログイン成功させるチートが可能になる
	// 出品・購入はしない
	wg.Add(1)
	go func() {
		defer wg.Done()

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(8 * time.Second)

			err := irregularLoginWrongPassword(ctx, user3)
			if err != nil {
				critical.Add(err)
			}

			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// check scenario #2
	// - カテゴリをチェック
	// - ユーザをチェック
	// - エラー処理が除かれていないかの確認
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error
		var category asset.AppCategory
		var userIDs []int64
		var userID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(10 * time.Second)

			s1, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			category = asset.GetRandomRootCategory()
			err = checkNewCategoryItemsAndItems(ctx, s1, category.ID, 10, 15)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			// active seller ユーザページ全件確認
			userIDs = asset.GetRandomActiveSellerIDs(5)
			for _, userID = range userIDs {
				err = checkUserItemsAndItems(ctx, s1, userID, 5)
				if err != nil {
					critical.Add(err)
					return
				}
			}

			// no active seller ユーザページ確認
			err = checkUserItemsAndItems(ctx, s1, s2.UserID, 0)
			if err != nil {
				critical.Add(err)
				goto Final
			}
			err = checkUserItemsAndItems(ctx, s2, s1.UserID, 0)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = irregularSellAndBuy(ctx, s1, s2, user3)
			if err != nil {
				critical.Add(err)
			}

			BuyerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// check scenario #3
	// bumpしてから新着をチェックする
	// TODO: 新着はbumpが新着に出ていることを確認してから、初期データを後ろの方までいい感じに遡りたい
	// TODO: 速度が上がるとbumpしたものが新着に無くなる可能性があるので、created_at的になければ更に遡るようにする
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2, s3 *session.Session
		var err error

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			// bumpは投稿した直後だとできないので必ず新しいユーザーでやる
			user1 := asset.GetRandomActiveSeller()
			s1, err = loginedSession(ctx, user1)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = checkBumpAndNewItems(ctx, s1, s2)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			s3, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = checkTransactionEvidence(ctx, s3, 3, 20)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			err = checkTransactionEvidence(ctx, s3, 10, 20)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	// check scenario #4
	// 出品した商品を編集する（100円を110円とかにする）
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error
		var price int
		var targetItemID int64

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(10 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				critical.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			targetItemID, err = sell(ctx, s1, price)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			// 打った商品探す

			err = itemEditNewItemWithLoginedSession(ctx, s1, targetItemID, price+10)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			err = buyComplete(ctx, s1, s2, targetItemID, price+10)
			if err != nil {
				critical.Add(err)

				goto Final
			}

			ActiveSellerPool.Enqueue(s1)
			BuyerPool.Enqueue(s2)

		Final:
			select {
			case <-ch:
			case <-ctx.Done():
				break L
			}
		}
	}()

	go func() {
		wg.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-ctx.Done():
	}
}

func checkBumpAndNewItems(ctx context.Context, s1, s2 *session.Session) error {
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
	var createdAt int64
	found := false
	for _, item := range items {
		if createdAt > 0 && createdAt < item.CreatedAt {
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

	createdAt = targetItemCreatedAt
	for _, item := range items {
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

// カテゴリページの商品をたどる
func checkNewCategoryItemsAndItems(ctx context.Context, s *session.Session, categoryID int, maxPage int64, checkItem int) error {
	category, ok := asset.GetCategory(categoryID)
	if !ok || category.ParentID != 0 {
		// benchmarkerのバグになるかと
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json カテゴリIDが正しくありません", categoryID))
	}
	itemIDs := newIDsStore()
	err := checkItemIDsFromCategory(ctx, s, itemIDs, categoryID, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// 全件チェックの時だけチェック
	// countUserItemsでもチェックしているので、商品数が最低数あればよい
	if (maxPage == 0 && c < 3000) || c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json の商品数が正しくありません", categoryID))
	}

	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := checkGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkItemIDsFromCategory(ctx context.Context, s *session.Session, itemIDs *IDsStore, categoryID int, nextItemID, nextCreatedAt, loop, maxPage int64) error {
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
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonはcreated_at順である必要があります", categoryID))
		}

		if item.Category.ParentID != categoryID {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json のカテゴリが異なります (item_id: %d)", categoryID, item.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			// 見つからない
			continue
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
		err := checkItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkTransactionEvidence(ctx context.Context, s *session.Session, maxPage int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := checkItemIDsTransactionEvidence(ctx, s, itemIDs, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	if c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品数が正しくありません (user_id: %d)", s.UserID))
	}
	if checkItem == 0 {
		return nil
	}
	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := checkGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkItemIDsTransactionEvidence(ctx context.Context, s *session.Session, itemIDs *IDsStore, nextItemID, nextCreatedAt, loop, maxPage int64) error {
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

	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonはcreated_at順である必要があります"))
		}

		if item.BuyerID != s.UserID && item.Seller.ID != s.UserID {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに購入・出品していない商品が含まれます (item_id: %d, user_id: %d)", item.ID, s.UserID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			// 見つからない
			continue
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの商品の名前が間違えています (item_id: %d)", item.ID))
		}

		err := checkItemDetailCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの%s", err.Error()))
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
		err := checkItemIDsTransactionEvidence(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil
}

// ユーザページをたどる
func checkUserItemsAndItems(ctx context.Context, s *session.Session, sellerID int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := checkItemIDsFromUsers(ctx, s, itemIDs, sellerID, 0, 0, 0)
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
		err := checkGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkItemIDsFromUsers(ctx context.Context, s *session.Session, itemIDs *IDsStore, sellerID, nextItemID, nextCreatedAt, loop int64) error {
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
		err := checkItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkGetItem(ctx context.Context, s *session.Session, targetItemID int64) error {
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
		// 見つからない
		return nil
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

	return nil
}
