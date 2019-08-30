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

func load(ctx context.Context, critical *fails.Critical) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	// load scenario #1
	// カテゴリを少しみてbuy
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var categories []asset.AppCategory
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

				categories = asset.GetRootCategories()
				for _, category := range categories {
					err = loadNewCategoryItemsAndItems(ctx, s1, category.ID, 20, 15)
					if err != nil {
						critical.Add(err)
						goto Final
					}
				}

				err = loadSellNewCategoryBuyWithLoginedSession(ctx, s1, s2, price)
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
	// どちらかというとカテゴリを中心にみていく
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var targetItemID int64
			var item *session.ItemDetail
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

				err = loadNewCategoryItemsAndItems(ctx, s2, item.Category.ParentID, 20, 5)
				if err != nil {
					critical.Add(err)
					goto Final
				}

				err = transactionEvidence(ctx, s1)
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
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s1, s2, s3 *session.Session
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
				err = loadUserItemsAndItems(ctx, s2, s1.UserID, 10)
				if err != nil {
					critical.Add(err)
					goto Final
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
	for i := 0; i < 3; i++ {
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
	var createdAt int64
	for _, item := range items {
		if createdAt > 0 && createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item.jsonはcreated_at順である必要があります"))
		}
		createdAt = item.CreatedAt

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
	var createdAt int64
	for _, item := range items {
		if createdAt > 0 && createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonはcreated_at順である必要があります", categoryID))
		}
		createdAt = item.CreatedAt

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
	var createdAt int64
	for _, item := range items {
		if createdAt > 0 && createdAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonはcreated_at順である必要があります", sellerID))
		}
		createdAt = item.CreatedAt

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
