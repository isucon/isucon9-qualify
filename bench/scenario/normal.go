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

const (
	MinCampaignRateSetting = 0
	MaxCampaignRateSetting = 4
	loadIDsMaxloop         = 100
)

func initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (int, string, error) {
	s1, err := session.NewSessionForInialize()
	if err != nil {
		return 0, "", err
	}
	campaign, language, err := s1.Initialize(ctx, paymentServiceURL, shipmentServiceURL)
	if err != nil {
		return 0, "", err
	}
	if campaign < MinCampaignRateSetting || campaign > MaxCampaignRateSetting {
		return 0, "", failure.New(fails.ErrApplication, failure.Messagef("POST /initialize の還元率の設定値は %d以上 %d以下です", MinCampaignRateSetting, MaxCampaignRateSetting))
	}

	if language == "" {
		return 0, "", failure.New(fails.ErrApplication, failure.Message("POST /initialize では実装言語を返す必要があります"))
	}

	return campaign, language, nil
}

func checkItemSimpleCategory(item session.ItemSimple, aItem asset.AppItem) error {
	aCategory, _ := asset.GetCategory(aItem.CategoryID)
	aRootCategory, _ := asset.GetCategory(aCategory.ParentID)

	if item.Category == nil {
		return fmt.Errorf("商品のカテゴリーがありません")
	}
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

	if item.Category == nil {
		return fmt.Errorf("商品のカテゴリーがありません")
	}
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

func findItemFromUsers(ctx context.Context, s *session.Session, targetItem asset.AppItem, maxPage int64) (session.ItemSimple, error) {
	return findItemFromUsersAll(ctx, s, targetItem.SellerID, targetItem.ID, 0, 0, 0, maxPage)
}

func findItemFromUsersByID(ctx context.Context, s *session.Session, sellerID, targetItemID, maxPage int64) (session.ItemSimple, error) {
	return findItemFromUsersAll(ctx, s, sellerID, targetItemID, 0, 0, 0, maxPage)
}

func findItemFromUsersAll(ctx context.Context, s *session.Session, sellerID, targetItemID, nextItemID, nextCreatedAt, loop, maxPage int64) (session.ItemSimple, error) {
	var hasNext bool
	var items []session.ItemSimple
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, _, items, err = s.UserItemsWithItemIDAndCreatedAt(ctx, sellerID, nextItemID, nextCreatedAt)
	} else {
		hasNext, _, items, err = s.UserItems(ctx, sellerID)
	}
	if err != nil {
		return session.ItemSimple{}, err
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/users/%d.jsonはcreated_at順である必要があります", sellerID))
		}
		if item.SellerID != sellerID {
			return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json の出品者が正しくありません　(item_id: %d)", sellerID, item.ID))
		}

		if item.ID == targetItemID {
			return item, nil
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json から商品を探すことができませんでした　(item_id: %d)", sellerID, targetItemID))
	}
	if hasNext && loop < loadIDsMaxloop {
		return findItemFromUsersAll(ctx, s, sellerID, targetItemID, nextItemID, nextCreatedAt, loop, maxPage)
	}
	return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/users/%d.json から商品を探すことができませんでした　(item_id: %d)", sellerID, targetItemID))
}

func findItemFromNewCategory(ctx context.Context, s *session.Session, targetItem asset.AppItem, maxPage int64) (session.ItemSimple, error) {
	targetCategory, ok := asset.GetCategory(targetItem.CategoryID)
	if !ok || targetCategory.ParentID == 0 {
		// データ不整合・ベンチマーカのバグの可能性
		return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("商品のカテゴリを探すことができませんでした (item_id: %d)", targetItem.ID))
	}
	targetRootCategoryID := targetCategory.ParentID
	return findItemFromNewCategoryAll(ctx, s, targetRootCategoryID, targetItem.ID, 0, 0, 0, maxPage)
}

func findItemFromNewCategoryAll(ctx context.Context, s *session.Session, categoryID int, targetItemID, nextItemID, nextCreatedAt, loop, maxPage int64) (session.ItemSimple, error) {
	var hasNext bool
	var items []session.ItemSimple
	var err error
	if nextItemID > 0 && nextCreatedAt > 0 {
		hasNext, _, items, err = s.NewCategoryItemsWithItemIDAndCreatedAt(ctx, categoryID, nextItemID, nextCreatedAt)
	} else {
		hasNext, _, items, err = s.NewCategoryItems(ctx, categoryID)
	}
	if err != nil {
		return session.ItemSimple{}, err
	}
	for _, item := range items {
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.jsonはcreated_at順である必要があります", categoryID))
		}
		if item.Category == nil {
			return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json のカテゴリが返っていません (item_id: %d)", categoryID, item.ID))
		}

		if item.Category.ParentID != categoryID {
			return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json のカテゴリが異なります (item_id: %d)", categoryID, item.ID))
		}

		if item.ID == targetItemID {
			return item, nil
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json から商品を探すことができませんでした　(item_id: %d)", categoryID, targetItemID))
	}
	if hasNext && loop < loadIDsMaxloop {
		return findItemFromNewCategoryAll(ctx, s, categoryID, targetItemID, nextItemID, nextCreatedAt, loop, maxPage)
	}
	return session.ItemSimple{}, failure.New(fails.ErrApplication, failure.Messagef("/new_items/%d.json から商品を探すことができませんでした　(item_id: %d)", categoryID, targetItemID))
}

func findItemFromUsersTransactions(ctx context.Context, s *session.Session, targetItemID, maxPage int64) (session.ItemDetail, error) {
	return findItemFromUsersTransactionsAll(ctx, s, targetItemID, 0, 0, 0, maxPage)
}

func findItemFromUsersTransactionsAll(ctx context.Context, s *session.Session, targetItemID, nextItemID, nextCreatedAt, loop, maxPage int64) (session.ItemDetail, error) {
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
		if nextCreatedAt > 0 && nextCreatedAt < item.CreatedAt {
			return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonはcreated_at順である必要があります"))
		}

		if item.Seller == nil {
			return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonの購入者情報が返っていません (item_id: %d, user_id: %d)", item.ID, s.UserID))
		}

		if item.BuyerID != s.UserID && item.Seller.ID != s.UserID {
			return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.jsonに購入・出品していない商品が含まれます (item_id: %d, user_id: %d)", item.ID, s.UserID))
		}

		if item.ID == targetItemID {
			return item, nil
		}
		nextItemID = item.ID
		nextCreatedAt = item.CreatedAt
	}
	loop = loop + 1
	if maxPage > 0 && loop >= maxPage {
		return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json から商品を探すことができませんでした　(item_id: %d)", targetItemID))
	}
	if hasNext && loop < loadIDsMaxloop {
		return findItemFromUsersTransactionsAll(ctx, s, targetItemID, nextItemID, nextCreatedAt, loop, maxPage)
	}
	return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json から商品を探すことができませんでした　(item_id: %d)", targetItemID))
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
