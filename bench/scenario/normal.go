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
	// TODO any check?
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
