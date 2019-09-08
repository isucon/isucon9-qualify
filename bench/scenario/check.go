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

func Check(ctx context.Context) {
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
		for j := 0; j < ExecutionSeconds/8; j++ {
			ch := time.After(8 * time.Second)

			err := irregularLoginWrongPassword(ctx, user3)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
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
		for j := 0; j < ExecutionSeconds/10; j++ {
			ch := time.After(10 * time.Second)

			s1, err = buyerSession(ctx)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			category = asset.GetRandomRootCategory()
			err = checkNewCategoryItemsAndItems(ctx, s1, category.ID, 10, 15)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			// active seller ユーザページ全件確認
			userIDs = asset.GetRandomActiveSellerIDs(5)
			for _, userID = range userIDs {
				err = checkUserItemsAndItems(ctx, s1, userID, 5)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					return
				}
			}

			// no active seller ユーザページ確認
			err = checkUserItemsAndItems(ctx, s1, s2.UserID, 0)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}
			err = checkUserItemsAndItems(ctx, s2, s1.UserID, 0)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			err = irregularSellAndBuy(ctx, s1, s2, user3)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
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
	// bumpしてからカテゴリ新着をチェックする
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error

	L:
		for j := 0; j < ExecutionSeconds/5; j++ {
			ch := time.After(5 * time.Second)

			// bumpは投稿した直後だとできないので必ず新しいユーザーでやる
			user1 := asset.GetRandomActiveSeller()
			s1, err = loginedSession(ctx, user1)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			err = checkBumpAndNewItems(ctx, s1, s2)
			if err != nil {
				fails.ErrorsForCheck.Add(err)

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
	// 出品した商品を探す
	// さいごは購入
	wg.Add(1)
	go func() {
		defer wg.Done()

		var s1, s2 *session.Session
		var err error
		var price, numSellBefore int
		var targetItem asset.AppItem
		var findItem session.ItemSimple
		var targetParentCategoryID int

	L:
		for j := 0; j < ExecutionSeconds/10; j++ {
			ch := time.After(10 * time.Second)

			s1, err = activeSellerSession(ctx)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			s2, err = buyerSession(ctx)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			price = priceStoreCache.Get()

			numSellBefore = asset.GetUser(s1.UserID).NumSellItems
			targetParentCategoryID = asset.GetUser(s2.UserID).BuyParentCategoryID
			targetItem, err = sellParentCategory(ctx, s1, price, targetParentCategoryID)
			if err != nil {
				fails.ErrorsForCheck.Add(err)

				goto Final
			}

			// 売った商品探す
			findItem, err = findItemFromUsers(ctx, s1, targetItem, 2)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}
			if !(findItem.Seller.NumSellItems > numSellBefore) {
				fails.ErrorsForCheck.Add(failure.New(fails.ErrApplication, failure.Messagef("ユーザの出品数が更新されていません (user_id:%d)", s1.UserID)))
				goto Final
			}
			_, err = findItemFromNewCategory(ctx, s1, targetItem, 3)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}
			_, err = findItemFromUsersTransactions(ctx, s1, targetItem.ID, 5)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			err = itemEditNewItemWithLoginedSession(ctx, s1, targetItem.ID, price+10)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
				goto Final
			}

			err = buyCompleteWithVerify(ctx, s1, s2, targetItem.ID, price+10)
			if err != nil {
				fails.ErrorsForCheck.Add(err)
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

	targetItem := asset.SetItemCreatedAt(s1.UserID, targetItemID, newCreatedAt)

	itemFromNewCategory, err := findItemFromNewCategory(ctx, s1, targetItem, 3)
	if err != nil {
		return err
	}
	if itemFromNewCategory.CreatedAt != newCreatedAt {
		return failure.New(fails.ErrApplication, failure.Messagef("Bump後の商品が更新されていません (item_id: %d)", targetItemID))
	}
	itemFromUsers, err := findItemFromUsers(ctx, s1, targetItem, 3)
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

	err = checkNewCategoryItemsAndItems(ctx, s1, targetCategory.ParentID, 2, 5)
	if err != nil {
		return err
	}

	return nil
}

// カテゴリページの商品をたどる
func checkNewCategoryItemsAndItems(ctx context.Context, s *session.Session, categoryID int, maxPage int64, checkItem int) error {
	category, ok := asset.GetCategory(categoryID)
	if !ok || category.ParentID != 0 {
		// benchmarkerのバグになるかと
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json カテゴリIDが正しくありません", categoryID))
	}
	itemIDs := newIDsStore()
	err := checkItemIDsFromCategory(ctx, s, itemIDs, categoryID, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// 全件はカウントできない。countUserItemsを何回か動かして確認している
	// ここでは商品数はperpage*maxpage
	if maxPage > 0 && int64(c) != maxPage*asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json の商品数が正しくありません", categoryID))
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
	if loop < 50 && asset.ItemsPerPage != len(items) { // MEMO 50件よりはみないだろう
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json の商品数が正しくありません", categoryID))
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonはcreated_at順である必要があります", categoryID))
		}

		if item.Category == nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json のカテゴリが異なります (item_id: %d)", categoryID, item.ID))
		}
		if item.Category.ParentID != categoryID {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json のカテゴリが異なります (item_id: %d)", categoryID, item.ID))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json の商品のステータスが正しくありません (item_id: %d)", categoryID, item.ID))
		}

		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			// 見つからない
			continue
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの商品の名前が間違えています (item_id: %d)", categoryID, item.ID))
		}

		err := checkItemSimpleCategory(item, aItem)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonの%s", categoryID, err.Error()))
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonに同じ商品がありました (item_id: %d)", categoryID, item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return nil
	}
	if hasNext && loop < loadIDsMaxloop {
		return checkItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
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
	buffer := 10 // 多少のずれは許容
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
	// 件数チェックはしない。合計でみている
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonはcreated_at順である必要があります", sellerID))
		}

		if item.SellerID != sellerID {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の出品者が正しくありません　(item_id: %d)", sellerID, item.ID))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut && item.Status != asset.ItemStatusTrading {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の商品のステータスが正しくありません (item_id: %d)", sellerID, item.ID))
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
	if hasNext && loop < loadIDsMaxloop {
		return checkItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
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

	if item.Seller == nil {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.json の商品の出品者情報が返っていません", targetItemID))
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
