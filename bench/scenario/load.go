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

const (
	NumLoadScenario1 = 3
	NumLoadScenario2 = 6
	NumLoadScenario3 = 6
	NumLoadScenario4 = 3
)

func Load(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	// load scenario #1
	// カテゴリを少しみてbuy
	for i := 0; i < NumLoadScenario1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var categories []asset.AppCategory
			var targetItemID int64
		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				categories = asset.GetRootCategories()
				for _, category := range categories {
					err = loadNewCategoryItemsAndItems(ctx, s1, category.ID, 20, 20)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
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
	}

	// load scenario #2
	// 出品 => そのカテゴリ => getTransactions => buy
	for i := 0; i < NumLoadScenario2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var targetItemID int64
			var item session.ItemDetail
		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				item, err = s1.Item(ctx, targetItemID)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = loadNewCategoryItemsAndItems(ctx, s1, item.Category.ParentID, 30, 20)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s1, 10, 20)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s2, 0, 0)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s1, 10, 20)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				// ここは厳密なcheckをしない
				err = buyComplete(ctx, s1, s2, targetItemID, price)
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
	}

	// load scenario #3
	// どちらかというとuserを中心にみていく
	for i := 0; i < NumLoadScenario3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s1, s2, s3 *session.Session
			var err error
			var price int
			var targetItemID int64
			var userIDs []int64

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				s3, err = buyerSession(ctx)
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

				// ユーザのページを全部みる。
				// activeユーザ3ページ
				userIDs = asset.GetRandomActiveSellerIDs(3)
				for _, userID := range userIDs {
					err = loadUserItemsAndItems(ctx, s2, userID, 20)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				// 商品数がすくないところもみにいく
				// indexつけるだけで速くなる
				for l := 0; l < 4; l++ {
					err = loadUserItemsAndItems(ctx, s1, s3.UserID, 0)
					if err != nil {
						critical.Add(err)
						goto Final
					}
					err = loadUserItemsAndItems(ctx, s3, s2.UserID, 0)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				ActiveSellerPool.Enqueue(s1)
				BuyerPool.Enqueue(s2)
				BuyerPool.Enqueue(s3)

			Final:
				select {
				case <-ch:
				case <-ctx.Done():
					break L
				}
			}
		}()
	}

	// load scenario #4
	// NewItemみてbuy
	for i := 0; i < NumLoadScenario4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s1, s2 *session.Session
			var err error
			var price int
			var targetItemID int64

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				err = loadNewItemsAndItems(ctx, s2, 30, 50)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItemID, price)
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
	}

	go func() {
		wg.Wait()
		close(closed)
	}()

	select {
	case <-closed:
	case <-ctx.Done():
	}
}

// Timelineの商品をたどる
func loadNewItemsAndItems(ctx context.Context, s *session.Session, maxPage int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := loadItemIDsFromNewItems(ctx, s, itemIDs, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// 全件チェックの時だけチェック
	// countUserItemsでもチェックしているので、商品数が最低数あればよい
	if (maxPage == 0 && c < 30000) || c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item.json の商品数が正しくありません"))
	}

	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := loadGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadItemIDsFromNewItems(ctx context.Context, s *session.Session, itemIDs *IDsStore, nextItemID, nextCreatedAt, loop, maxPage int64) error {
	var hasNext bool
	var items []session.ItemSimple
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, items, err = s.NewItemsWithItemIDAndCreatedAt(ctx, nextItemID, nextCreatedAt)
	} else {
		hasNext, items, err = s.NewItems(ctx)
	}
	if err != nil {
		return err
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item.jsonはcreated_at順である必要があります"))
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item.jsonに同じ商品がありました (item_id: %d)", item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return nil
	}
	if hasNext && loop < 100 { // TODO: max pager
		err := loadItemIDsFromNewItems(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil

}

// カテゴリページの商品をたどる
func loadNewCategoryItemsAndItems(ctx context.Context, s *session.Session, categoryID int, maxPage int64, checkItem int) error {
	category, ok := asset.GetCategory(categoryID)
	if !ok || category.ParentID != 0 {
		// benchmarkerのバグになるかと
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json カテゴリIDが正しくありません", categoryID))
	}
	itemIDs := newIDsStore()
	err := loadItemIDsFromCategory(ctx, s, itemIDs, categoryID, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	/*
		mysql> alter table items add root_category_id int unsigned NOT NULL after category_id;
		mysql> update items i join categories c on i.category_id=c.id set i.root_category_id = c.parent_id;
		mysql> select root_category_id,count(*) from items group by root_category_id;
		+------------------+----------+
		| root_category_id | count(*) |
		+------------------+----------+
		|                1 |     3886 |
		|               10 |     7192 |
		|               20 |     8501 |
		|               30 |     7461 |
		|               40 |     7734 |
		|               50 |     5125 |
		|               60 |    10395 |
		+------------------+----------+
		7 rows in set (0.04 sec)
	*/
	// 全件チェックの時だけチェック
	// countUserItemsでもチェックしているので、商品数が最低数あればよい
	if (maxPage == 0 && c < 3000) || c < checkItem {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.json の商品数が正しくありません", categoryID))
	}

	chkItemIDs := itemIDs.RandomIDs(checkItem)
	for _, itemID := range chkItemIDs {
		err := loadGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}

	return nil
}

func loadItemIDsFromCategory(ctx context.Context, s *session.Session, itemIDs *IDsStore, categoryID int, nextItemID, nextCreatedAt, loop, maxPage int64) error {
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
		err := loadItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil
}

// ユーザページをたどる
func loadUserItemsAndItems(ctx context.Context, s *session.Session, sellerID int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := loadItemIDsFromUsers(ctx, s, itemIDs, sellerID, 0, 0, 0)
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
		err := loadGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadItemIDsFromUsers(ctx context.Context, s *session.Session, itemIDs *IDsStore, sellerID, nextItemID, nextCreatedAt, loop int64) error {
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

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonに同じ商品がありました (item_id: %d)", sellerID, item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if hasNext && loop < 100 { // TODO: max pager
		err := loadItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadTransactionEvidence(ctx context.Context, s *session.Session, maxPage int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := loadItemIDsTransactionEvidence(ctx, s, itemIDs, 0, 0, 0, maxPage)
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
		err := loadGetItem(ctx, s, itemID)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadItemIDsTransactionEvidence(ctx context.Context, s *session.Session, itemIDs *IDsStore, nextItemID, nextCreatedAt, loop, maxPage int64) error {
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
		err := loadItemIDsTransactionEvidence(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadGetItem(ctx context.Context, s *session.Session, targetItemID int64) error {
	item, err := s.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	if !(item.Description != "") {
		return failure.New(fails.ErrApplication, failure.Messagef("/items/%d.jsonの商品説明が間違っています", targetItemID))
	}

	return nil
}
