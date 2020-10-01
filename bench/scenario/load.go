package scenario

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	// シナリオ(1,2,3,4) = 並列数(1,2,2,1)
	// これを負荷の1単位とする
	// 1だとLoad内のfor loopが必要ないが、調整のため残す
	NumLoadScenario1 = 1
	NumLoadScenario2 = 2
	NumLoadScenario3 = 2
	NumLoadScenario4 = 1
)

func Load(ctx context.Context) {
	var wg sync.WaitGroup
	closed := make(chan struct{})

	// 以下の関数はすべてsellとbuyの間に他の処理を挟む
	// 今回の問題は決済総額がスコアになるのでMySQLを守るためにGETの速度を落とすチートが可能
	// それを防ぐためにsellしたあとに他のエンドポイントにリクエストを飛ばして完了してからbuyされる
	// シナリオとしてはGETで色んなページを見てから初めて購入に結びつくという動きをするのは自然
	// 最適化が難しいエンドポイントの速度をわざと落として、最適化が簡単なエンドポイントに負荷を偏らせるチートを防ぐために
	// すべてのシナリオはチャネルを使って一定時間より早く再実行はしないようにする
	// 理論上そのエンドポイントを高速化することで出せるスコアに上限が出るので、他のエンドポイントを最適化する必要性が出る

	// load scenario #1
	// 出品
	// カテゴリをみて 7カテゴリ x (10ページ + 20item) = 210
	// recommendであれば、Newだけみて、購入し、再度出品・購入がある
	// buy without check
	for i := 0; i < NumLoadScenario1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2, s3 *session.Session
			var err error
			var price int
			var categories []asset.AppCategory
			var targetItem asset.AppItem
			var recommended bool
			var targetParentCategoryID int

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				s3, err = activeSellerSession(ctx)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				recommended, err = loadIsRecommendNewItems(ctx, s2)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				price = priceStoreCache.Get()

				targetParentCategoryID = asset.GetUser(s2.UserID).BuyParentCategoryID
				targetItem, err = sellParentCategory(ctx, s1, price, targetParentCategoryID)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				if recommended {
					// recommended なら categoryは見ずにnewをみる
					err = loadNewItemsAndItems(ctx, s2, 10, 20)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}
				} else {
					categories = asset.GetRootCategories()
					for _, category := range categories {
						err = loadNewCategoryItemsAndItems(ctx, s2, category.ID, 10, 20)
						if err != nil {
							fails.ErrorsForCheck.Add(err)
							goto Final
						}
					}
				}

				err = buyComplete(ctx, s1, s2, targetItem.ID, price)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				// recommended なら購入2倍
				if recommended {
					targetItem, err = sellParentCategory(ctx, s3, price, targetParentCategoryID)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}

					// 少しだけNewItemをみて購入
					err = loadNewItemsAndItems(ctx, s2, 1, 10)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}

					err = buyComplete(ctx, s3, s2, targetItem.ID, price)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}
				}

				ActiveSellerPool.Enqueue(s1)
				BuyerPool.Enqueue(s2)
				ActiveSellerPool.Enqueue(s3)

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
	// 出品
	// その商品
	// そのカテゴリ 30ページ 30商品
	// getTransactions　(10ページ 20商品) x 2
	// buyはwithout check
	for i := 0; i < NumLoadScenario2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var s1, s2 *session.Session
			var err error
			var price int
			var targetItem asset.AppItem
			var item session.ItemDetail
			var targetParentCategoryID int

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				targetParentCategoryID = asset.GetUser(s2.UserID).BuyParentCategoryID
				targetItem, err = sellParentCategory(ctx, s1, price, targetParentCategoryID)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				item, err = s1.Item(ctx, targetItem.ID)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				if item.Category == nil {
					fails.ErrorsForCheck.Add(failure.New(fails.ErrApplication, failure.Messagef("/item/%d.json のカテゴリが正しくありません", item.ID)))
					goto Final
				}

				err = loadNewCategoryItemsAndItems(ctx, s1, item.Category.ParentID, 30, 20)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s1, 10, 20)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s2, 0, 0)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = loadTransactionEvidence(ctx, s1, 10, 20)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = buyComplete(ctx, s1, s2, targetItem.ID, price)
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
	}

	// load scenario #3
	// どちらかというとuserを中心にみていく
	// 出品
	// アクティブユーザ 3人 * (3ページ + 20件)
	// buy with check
	for i := 0; i < NumLoadScenario3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s1, s2, s3 *session.Session
			var err error
			var price int
			var targetItem asset.AppItem
			var userIDs []int64
			var targetParentCategoryID int

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				s3, err = buyerSession(ctx)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				price = priceStoreCache.Get()

				targetParentCategoryID = asset.GetUser(s2.UserID).BuyParentCategoryID
				targetItem, err = sellParentCategory(ctx, s1, price, targetParentCategoryID)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				// ユーザのページを全部みる。
				// activeユーザ3ページ
				userIDs = asset.GetRandomActiveSellerIDs(3)
				for _, userID := range userIDs {
					err = loadUserItemsAndItems(ctx, s2, userID, 20)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}
				}

				// 商品数がすくないところもみにいく
				// indexつけるだけで速くなる
				for l := 0; l < 4; l++ {
					err = loadUserItemsAndItems(ctx, s1, s3.UserID, 0)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}
					err = loadUserItemsAndItems(ctx, s3, s2.UserID, 0)
					if err != nil {
						fails.ErrorsForCheck.Add(err)
						goto Final
					}
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItem.ID, price)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
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
	// 出品
	// 新着 30ページ 50商品
	// buy with check
	for i := 0; i < NumLoadScenario4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var s1, s2 *session.Session
			var err error
			var price int
			var targetItem asset.AppItem
			var targetParentCategoryID int

		L:
			for j := 0; j < ExecutionSeconds/3; j++ {
				ch := time.After(3 * time.Second)

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

				targetParentCategoryID = asset.GetUser(s2.UserID).BuyParentCategoryID
				targetItem, err = sellParentCategory(ctx, s1, price, targetParentCategoryID)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = loadNewItemsAndItems(ctx, s2, 30, 50)
				if err != nil {
					fails.ErrorsForCheck.Add(err)
					goto Final
				}

				err = buyCompleteWithVerify(ctx, s1, s2, targetItem.ID, price)
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

// Timeline が recommend になっているか
func loadIsRecommendNewItems(ctx context.Context, s *session.Session) (bool, error) {
	aUser := asset.GetUser(s.UserID)
	targetCategoryID := aUser.BuyParentCategoryID

	_, items, err := s.NewItems(ctx)
	if err != nil {
		return false, err
	}
	if len(items) != asset.ItemsPerPage {
		return false, failure.New(fails.ErrApplication, failure.Messagef("/new_items.json の商品数が正しくありません"))
	}
	isTarget := 0
	for _, item := range items {
		if item.Category == nil {
			return false, failure.New(fails.ErrApplication, failure.Messagef("/new_items.json のカテゴリが正しくありません　(item_id: %d)", item.ID))
		}
		if item.Category.ParentID == targetCategoryID {
			isTarget++
		}
	}
	if float64(isTarget)/float64(len(items)) > 0.8 {
		return true, nil
	}
	return false, nil
}

// Timelineの商品をたどる
func loadNewItemsAndItems(ctx context.Context, s *session.Session, maxPage int64, checkItem int) error {
	itemIDs := newIDsStore()
	err := loadItemIDsFromNewItems(ctx, s, itemIDs, 0, 0, 0, maxPage)
	if err != nil {
		return err
	}
	c := itemIDs.Len()
	// 全件はカウントできない。countUserItemsを何回か動かして確認している
	// ここでは商品数はperpage*maxpage
	if maxPage > 0 && int64(c) != maxPage*asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items.json の商品数が正しくありません"))
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
	if loop < 50 && asset.ItemsPerPage != len(items) { // MEMO 50件よりはみないだろう
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items.json の商品数が正しくありません"))
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonはcreated_at順である必要があります"))
		}

		if item.Status != asset.ItemStatusOnSale && item.Status != asset.ItemStatusSoldOut {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.json の商品のステータスが正しくありません (item_id: %d)", item.ID))
		}

		err = itemIDs.Add(item.ID)
		if err != nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_items.jsonに同じ商品がありました (item_id: %d)", item.ID))
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return nil
	}
	if hasNext && loop < loadIDsMaxloop {
		return loadItemIDsFromNewItems(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
	}
	return nil

}

// カテゴリページの商品をたどる
func loadNewCategoryItemsAndItems(ctx context.Context, s *session.Session, categoryID int, maxPage int64, checkItem int) error {
	category, ok := asset.GetCategory(categoryID)
	if !ok || category.ParentID != 0 {
		// benchmarkerのバグになるかと
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json カテゴリIDが正しくありません", categoryID))
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
	// 全件はカウントできない。countUserItemsを何回か動かして確認している
	// ここでは商品数はperpage*maxpage
	if maxPage > 0 && int64(c) != maxPage*asset.ItemsPerPage {
		return failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json の商品数が正しくありません", categoryID))
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
	if loop < 50 && len(items) != asset.ItemsPerPage { // MEMO 50ページ以上チェックすることはない
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
		return loadItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
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
	// 件数のチェックはない。userは全部みて件数確認する
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
	if hasNext && loop < loadIDsMaxloop {
		return loadItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
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
	aUser := asset.GetUser(s.UserID)
	totalTrxItems := aUser.NumBuyItems + aUser.NumSellItems
	maxPageItems := maxPage * asset.ItemsTransactionsPerPage
	if maxPage == 0 {
		maxPageItems = asset.ItemsTransactionsPerPage
	}
	// totalTrxItemsが多い場合 c は maxPageItems になる。5個のずれは許容
	if int64(totalTrxItems) >= maxPageItems && math.Abs(float64(c)-float64(maxPageItems)) > 5 {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品数が正しくありません (user_id: %d)", s.UserID))
	}
	// totalTrxItems が少ない場合、 cはtotalTrxItemsになる。5個のずれは許容
	if int64(totalTrxItems) < maxPageItems && math.Abs(float64(c)-float64(totalTrxItems)) > 5 {
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
	if hasNext && asset.ItemsTransactionsPerPage != len(items) {
		return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品数が正しくありません (user_id: %d)", s.UserID))
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonはcreated_at順である必要があります"))
		}

		if item.Seller == nil {
			return failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json の商品の出品者情報が返っていません (item_id: %d, user_id: %d)", item.ID, s.UserID))
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
	if hasNext && loop < loadIDsMaxloop {
		return loadItemIDsTransactionEvidence(ctx, s, itemIDs, nextItemID, nextCreatedAt, loop, maxPage)
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
