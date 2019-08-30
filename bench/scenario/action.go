package scenario

import (
	"context"
	"sync"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/isucon/isucon9-qualify/bench/server"
	"github.com/isucon/isucon9-qualify/bench/session"
	"github.com/morikuni/failure"
)

const (
	CorrectCardNumber = "AAAAAAAA"
	FailedCardNumber  = "FA10AAAA"
)

func activeSellerSession(ctx context.Context) (*session.Session, error) {
	s := ActiveSellerPool.Dequeue()
	if s != nil {
		return s, nil
	}

	user1 := asset.GetRandomActiveSeller()
	return loginedSession(ctx, user1)
}

func buyerSession(ctx context.Context) (*session.Session, error) {
	s := BuyerPool.Dequeue()
	if s != nil {
		return s, nil
	}

	user1 := asset.GetRandomBuyer()
	return loginedSession(ctx, user1)
}

func loginedSession(ctx context.Context, user1 asset.AppUser) (*session.Session, error) {
	s1, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	user, err := s1.Login(ctx, user1.AccountName, user1.Password)
	if err != nil {
		return nil, err
	}

	if !user1.Equal(user) {
		return nil, failure.New(fails.ErrApplication, failure.Message("ログインが失敗しています"))
	}

	err = s1.SetSettings(ctx)
	if err != nil {
		return nil, err
	}

	return s1, nil
}

func sell(ctx context.Context, s1 *session.Session, price int) (int64, error) {
	name, description, categoryID := asset.GenText(8, false), asset.GenText(200, true), 32

	targetItemID, err := s1.Sell(ctx, name, price, description, categoryID)
	if err != nil {
		return 0, err
	}

	asset.SetItem(s1.UserID, targetItemID, name, price, description, categoryID)

	return targetItemID, nil
}

func getItemIDsFromCategory(ctx context.Context, s *session.Session, itemIDs *IDsStore, categoryID int, nextItemID, nextCreatedAt, loop, maxPage int64) error {
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
		aItem, ok := asset.GetItem(item.SellerID, item.ID)
		if !ok {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonに存在しない商品 (item_id: %d) が返ってきています", categoryID, item.ID))
		}

		if !(item.Name == aItem.Name) {
			return failure.New(fails.ErrApplication, failure.Messagef("/new_item/%d.jsonの商品の名前が間違えています", categoryID))
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
		err := getItemIDsFromCategory(ctx, s, itemIDs, categoryID, nextItemID, nextCreatedAt, loop, maxPage)
		if err != nil {
			return err
		}
	}
	return nil

}

func getItemIDsFromUsers(ctx context.Context, s *session.Session, itemIDs *IDsStore, sellerID, nextItemID, nextCreatedAt, loop int64) error {
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
		err := getItemIDsFromUsers(ctx, s, itemIDs, sellerID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return err
		}
	}
	return nil
}

func FindItemFromUsersTransactions(ctx context.Context, s *session.Session, targetItemID int64) (session.ItemDetail, error) {
	return findItemFromUsersTransactions(ctx, s, targetItemID, 0, 0, 0)
}

func findItemFromUsersTransactions(ctx context.Context, s *session.Session, targetItemID, nextItemID, nextCreatedAt, loop int64) (session.ItemDetail, error) {
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
		_, err := findItemFromUsersTransactions(ctx, s, targetItemID, nextItemID, nextCreatedAt, loop)
		if err != nil {
			return session.ItemDetail{}, err
		}
	}
	return session.ItemDetail{}, failure.New(fails.ErrApplication, failure.Messagef("/users/transactions.json から商品を探すことができませんでした　(item_id: %d)", targetItemID))
}

func buyCompleteWithVerify(ctx context.Context, s1, s2 *session.Session, targetItemID int64, price int) error {
	token := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

	_, err := s2.Buy(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	itemFromBuyerTrx, err := FindItemFromUsersTransactions(ctx, s2, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err := FindItemFromUsersTransactions(ctx, s1, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyer, err := s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err := s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != asset.ItemStatusTrading || itemFromSeller.Status != asset.ItemStatusTrading ||
		itemFromBuyerTrx.Status != asset.ItemStatusTrading || itemFromSellerTrx.Status != asset.ItemStatusTrading {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromSeller.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromBuyerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromSellerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != asset.ShippingsStatusInitial ||
		itemFromSeller.ShippingStatus != asset.ShippingsStatusInitial ||
		itemFromBuyerTrx.ShippingStatus != asset.ShippingsStatusInitial ||
		itemFromSellerTrx.ShippingStatus != asset.ShippingsStatusInitial {
		return failure.New(fails.ErrApplication, failure.Messagef("購入後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	reserveID, apath, err := s1.Ship(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = FindItemFromUsersTransactions(ctx, s2, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err = FindItemFromUsersTransactions(ctx, s1, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != asset.ItemStatusTrading || itemFromSeller.Status != asset.ItemStatusTrading ||
		itemFromBuyerTrx.Status != asset.ItemStatusTrading || itemFromSellerTrx.Status != asset.ItemStatusTrading {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromSeller.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromBuyerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping ||
		itemFromSellerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitShipping {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != asset.ShippingsStatusWaitPickup ||
		itemFromSeller.ShippingStatus != asset.ShippingsStatusWaitPickup ||
		itemFromBuyerTrx.ShippingStatus != asset.ShippingsStatusWaitPickup ||
		itemFromSellerTrx.ShippingStatus != asset.ShippingsStatusWaitPickup {
		return failure.New(fails.ErrApplication, failure.Messagef("集荷予約後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	md5Str, err := s1.DownloadQRURL(ctx, apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(fails.ErrApplication, failure.Messagef("QRコードの画像に誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID))
	}

	err = s1.ShipDone(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemFromSellerTrx, err = FindItemFromUsersTransactions(ctx, s1, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = FindItemFromUsersTransactions(ctx, s2, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != asset.ItemStatusTrading || itemFromSeller.Status != asset.ItemStatusTrading ||
		itemFromBuyerTrx.Status != asset.ItemStatusTrading || itemFromSellerTrx.Status != asset.ItemStatusTrading {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitDone ||
		itemFromSeller.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitDone ||
		itemFromBuyerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitDone ||
		itemFromSellerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusWaitDone {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != asset.ShippingsStatusShipping ||
		itemFromSeller.ShippingStatus != asset.ShippingsStatusShipping ||
		itemFromBuyerTrx.ShippingStatus != asset.ShippingsStatusShipping ||
		itemFromSellerTrx.ShippingStatus != asset.ShippingsStatusShipping {
		return failure.New(fails.ErrApplication, failure.Messagef("発送完了後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("配送予約IDに誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID))
	}

	err = s2.Complete(ctx, targetItemID)
	if err != nil {
		return err
	}

	itemFromBuyer, err = s2.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromSellerTrx, err = FindItemFromUsersTransactions(ctx, s1, targetItemID)
	if err != nil {
		return err
	}
	itemFromSeller, err = s1.Item(ctx, targetItemID)
	if err != nil {
		return err
	}
	itemFromBuyerTrx, err = FindItemFromUsersTransactions(ctx, s2, targetItemID)
	if err != nil {
		return err
	}

	// status 確認
	if itemFromBuyer.Status != asset.ItemStatusSoldOut || itemFromSeller.Status != asset.ItemStatusSoldOut ||
		itemFromBuyerTrx.Status != asset.ItemStatusSoldOut || itemFromSellerTrx.Status != asset.ItemStatusSoldOut {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後の商品のステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.TransactionEvidenceStatus != asset.TransactionEvidenceStatusDone ||
		itemFromSeller.TransactionEvidenceStatus != asset.TransactionEvidenceStatusDone ||
		itemFromBuyerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusDone ||
		itemFromSellerTrx.TransactionEvidenceStatus != asset.TransactionEvidenceStatusDone {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後のtransaction_evidenceのステータスが正しくありません (item_id: %d)", targetItemID))
	}
	if itemFromBuyer.ShippingStatus != asset.ShippingsStatusDone ||
		itemFromSeller.ShippingStatus != asset.ShippingsStatusDone ||
		itemFromBuyerTrx.ShippingStatus != asset.ShippingsStatusDone ||
		itemFromSellerTrx.ShippingStatus != asset.ShippingsStatusDone {
		return failure.New(fails.ErrApplication, failure.Messagef("取引完了後のshippingのステータスが正しくありません (item_id: %d)", targetItemID))
	}

	return nil
}

func buyComplete(ctx context.Context, s1, s2 *session.Session, targetItemID int64, price int) error {
	token := sPayment.ForceSet(CorrectCardNumber, targetItemID, price)

	_, err := s2.Buy(ctx, targetItemID, token)
	if err != nil {
		return err
	}

	reserveID, apath, err := s1.Ship(ctx, targetItemID)
	if err != nil {
		return err
	}

	md5Str, err := s1.DownloadQRURL(ctx, apath)
	if err != nil {
		return err
	}

	sShipment.ForceSetStatus(reserveID, server.StatusShipping)
	if !sShipment.CheckQRMD5(reserveID, md5Str) {
		return failure.New(fails.ErrApplication, failure.Messagef("QRコードの画像に誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID))
	}

	err = s1.ShipDone(ctx, targetItemID)
	if err != nil {
		return err
	}

	ok := sShipment.ForceSetStatus(reserveID, server.StatusDone)
	if !ok {
		return failure.New(fails.ErrApplication, failure.Messagef("配送予約IDに誤りがあります (item_id: %d, reserve_id: %s)", targetItemID, reserveID))
	}

	err = s2.Complete(ctx, targetItemID)
	if err != nil {
		return err
	}

	return nil
}

type priceStore struct {
	price int
	sync.RWMutex
}

var priceStoreCache *priceStore

func init() {
	priceStoreCache = &priceStore{}
	priceStoreCache.Set(100)
}

func (s *priceStore) Get() int {
	s.RLock()
	defer s.RUnlock()
	return s.price
}

func (s *priceStore) Set(price int) {
	s.Lock()
	defer s.Unlock()
	s.price = price
}

func SetShipment(ss *server.ServerShipment) {
	sShipment = ss
}

func SetPayment(sp *server.ServerPayment) {
	sPayment = sp
}
