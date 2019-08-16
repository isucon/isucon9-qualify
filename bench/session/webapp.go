package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/tuotoo/qrcode"
	"golang.org/x/xerrors"
)

type User struct {
	ID             int64     `json:"id" db:"id"`
	AccountName    string    `json:"account_name" db:"account_name"`
	HashedPassword []byte    `json:"-" db:"hashed_password"`
	Address        string    `json:"address,omitempty" db:"address"`
	NumSellItems   int       `json:"num_sell_items" db:"num_sell_items"`
	LastBump       time.Time `json:"-" db:"last_bump"`
	CreatedAt      time.Time `json:"-" db:"created_at"`
}

type UserSimple struct {
	ID           int64  `json:"id"`
	AccountName  string `json:"account_name"`
	NumSellItems int    `json:"num_sell_items"`
}

type Item struct {
	ID          int64     `json:"id" db:"id"`
	SellerID    int64     `json:"seller_id" db:"seller_id"`
	BuyerID     int64     `json:"buyer_id" db:"buyer_id"`
	Status      string    `json:"status" db:"status"`
	Name        string    `json:"name" db:"name"`
	Price       int       `json:"price" db:"price"`
	Description string    `json:"description" db:"description"`
	CategoryID  int       `json:"category_id" db:"category_id"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type ItemSimple struct {
	ID         int64       `json:"id"`
	SellerID   int64       `json:"seller_id"`
	Seller     *UserSimple `json:"seller"`
	Status     string      `json:"status"`
	Name       string      `json:"name"`
	Price      int         `json:"price"`
	CategoryID int         `json:"category_id"`
	Category   *Category   `json:"category"`
	CreatedAt  int64       `json:"created_at"`
}

type ItemDetail struct {
	ID                        int64       `json:"id"`
	SellerID                  int64       `json:"seller_id"`
	Seller                    *UserSimple `json:"seller"`
	BuyerID                   int64       `json:"buyer_id,omitempty"`
	Buyer                     *UserSimple `json:"buyer,omitempty"`
	Status                    string      `json:"status"`
	Name                      string      `json:"name"`
	Price                     int         `json:"price"`
	Description               string      `json:"description"`
	CategoryID                int         `json:"category_id"`
	Category                  *Category   `json:"category"`
	TransactionEvidenceID     int64       `json:"transaction_evidence_id,omitempty"`
	TransactionEvidenceStatus string      `json:"transaction_evidence_status,omitempty"`
	ShippingStatus            string      `json:"shipping_status,omitempty"`
	CreatedAt                 int64       `json:"created_at"`
}

type TransactionEvidence struct {
	ID                 int64     `json:"id" db:"id"`
	SellerID           int64     `json:"seller_id" db:"seller_id"`
	BuyerID            int64     `json:"buyer_id" db:"buyer_id"`
	Status             string    `json:"status" db:"status"`
	ItemID             int64     `json:"item_id" db:"item_id"`
	ItemName           string    `json:"item_name" db:"item_name"`
	ItemPrice          int       `json:"item_price" db:"item_price"`
	ItemDescription    string    `json:"item_description" db:"item_description"`
	ItemCategoryID     int       `json:"item_category_id" db:"item_category_id"`
	ItemRootCategoryID int       `json:"item_root_category_id" db:"item_root_category_id"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"-" db:"updated_at"`
}

type Shipping struct {
	TransactionEvidenceID int64     `json:"transaction_evidence_id" db:"transaction_evidence_id"`
	Status                string    `json:"status" db:"status"`
	ItemName              string    `json:"item_name" db:"item_name"`
	ItemID                int64     `json:"item_id" db:"item_id"`
	ReserveID             string    `json:"reserve_id" db:"reserve_id"`
	ReserveTime           int64     `json:"reserve_time" db:"reserve_time"`
	ToAddress             string    `json:"to_address" db:"to_address"`
	ToName                string    `json:"to_name" db:"to_name"`
	FromAddress           string    `json:"from_address" db:"from_address"`
	FromName              string    `json:"from_name" db:"from_name"`
	ImgBinary             []byte    `json:"-" db:"img_binary"`
	CreatedAt             time.Time `json:"-" db:"created_at"`
	UpdatedAt             time.Time `json:"-" db:"updated_at"`
}

type Category struct {
	ID                 int    `json:"id" db:"id"`
	ParentID           int    `json:"parent_id" db:"parent_id"`
	CategoryName       string `json:"category_name" db:"category_name"`
	ParentCategoryName string `json:"parent_category_name,omitempty" db:"-"`
}

type resSetting struct {
	CSRFToken string `json:"csrf_token"`
}

type resSell struct {
	ID int64 `json:"id"`
}

type reqLogin struct {
	AccountName string `json:"account_name"`
	Password    string `json:"password"`
}

type reqItemEdit struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	ItemPrice int    `json:"item_price"`
}

type reqBuy struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	Token     string `json:"token"`
}

type resBuy struct {
	TransactionEvidenceID int64 `json:"transaction_evidence_id"`
}

type reqSell struct {
	CSRFToken   string `json:"csrf_token"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	CategoryID  int    `json:"category_id"`
}

type reqShip struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

type resShip struct {
	Path string `json:"path"`
}

type reqBump struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

type resItemEdit struct {
	ItemID        int64 `json:"item_id"`
	ItemPrice     int   `json:"item_price"`
	ItemCreatedAt int64 `json:"item_created_at"`
	ItemUpdatedAt int64 `json:"item_updated_at"`
}

type resNewItems struct {
	RootCategoryID   int          `json:"root_category_id,omitempty"`
	RootCategoryName string       `json:"root_category_name,omitempty"`
	HasNext          bool         `json:"has_next"`
	Items            []ItemSimple `json:"items"`
}

type resTransactions struct {
	HasNext bool         `json:"has_next"`
	Items   []ItemDetail `json:"items"`
}

type resUserItems struct {
	User    *UserSimple  `json:"user"`
	HasNext bool         `json:"has_next"`
	Items   []ItemSimple `json:"items"`
}

func (s *Session) Login(accountName, password string) (*asset.AppUser, error) {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, fails.NewError(err, "POST /login: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: "+msg)
	}

	u := &asset.AppUser{}
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, fails.NewError(err, "POST /login: JSONデコードに失敗しました")
	}

	return u, nil
}

func (s *Session) SetSettings() error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/settings")
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /settings: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /settings: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /settings: "+msg)
	}

	rs := &resSetting{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /settings: JSONデコードに失敗しました")
	}

	if rs.CSRFToken == "" {
		return fails.NewError(fmt.Errorf("csrf token is empty"), "GET /settings: csrf tokenが空でした")
	}

	s.csrfToken = rs.CSRFToken
	return nil
}

func (s *Session) Sell(name string, price int, description string, categoryID int) (int64, error) {
	b, _ := json.Marshal(reqSell{
		CSRFToken:   s.csrfToken,
		Name:        name,
		Price:       price,
		Description: description,
		CategoryID:  categoryID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: "+msg)
	}

	rs := &resSell{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /sell: JSONデコードに失敗しました")
	}

	return rs.ID, nil
}

func (s *Session) Buy(itemID int64, token string) (int64, error) {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: "+msg)
	}

	rb := &resBuy{}
	err = json.NewDecoder(res.Body).Decode(rb)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /buy: JSONデコードに失敗しました")
	}

	return rb.TransactionEvidenceID, nil
}

func (s *Session) Ship(itemID int64) (apath string, err error) {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: "+msg)
	}

	rs := &resShip{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return "", fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship: JSONデコードに失敗しました")
	}

	if len(rs.Path) == 0 {
		return "", fails.NewError(nil, "POST /ship: Pathが空です")
	}

	return rs.Path, nil
}

func (s *Session) ShipDone(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /ship_done: bodyの読み込みに失敗しました")
	}

	return nil
}

func (s *Session) Complete(itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/complete", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /complete: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /complete: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /complete: "+msg)
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /complete: bodyの読み込みに失敗しました")
	}

	return nil
}

func (s *Session) DecodeQRURL(apath string) (*url.URL, error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: リクエストに失敗しました", apath))
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: %s", apath, msg))
	}

	qrmatrix, err := qrcode.Decode(res.Body)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: QRコードがデコードできませんでした", apath))
	}

	surl := qrmatrix.Content

	if len(surl) == 0 {
		return nil, fails.NewError(nil, fmt.Sprintf("GET %s: QRコードの中身が空です", apath))
	}

	sparsedURL, err := url.ParseRequestURI(surl)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET %s: QRコードの中身がURLではありません", apath))
	}

	if sparsedURL.Host != ShareTargetURLs.ShipmentURL.Host {
		return nil, fails.NewError(nil, fmt.Sprintf("GET %s: shipment serviceのドメイン以外のURLがQRコードに表示されています", apath))
	}

	return sparsedURL, nil
}

func (s *Session) Bump(itemID int64) (int64, error) {
	b, _ := json.Marshal(reqBump{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/bump", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /bump: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /bump: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /bump: "+msg)
	}

	rie := &resItemEdit{}
	err = json.NewDecoder(res.Body).Decode(rie)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /bump: JSONデコードに失敗しました")
	}

	return rie.ItemCreatedAt, nil
}

func (s *Session) ItemEdit(itemID int64, price int) (int, error) {
	b, _ := json.Marshal(reqItemEdit{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		ItemPrice: price,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/items/edit", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /items/edit: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /items/edit: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /items/edit: "+msg)
	}

	rie := &resItemEdit{}
	err = json.NewDecoder(res.Body).Decode(rie)
	if err != nil {
		return 0, fails.NewError(xerrors.Errorf("error in session: %v", err), "POST /items/edit: JSONデコードに失敗しました")
	}

	return rie.ItemPrice, nil
}

func (s *Session) NewItems() (hasNext bool, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/new_items.json")
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: "+msg)
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: JSONデコードに失敗しました")
	}

	return rni.HasNext, rni.Items, nil
}

func (s *Session) NewItemsWithItemIDAndCreatedAt(itemID, createdAt int64) (hasNext bool, items []ItemSimple, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, "/new_items.json", q)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: "+msg)
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /new_items.json: JSONデコードに失敗しました")
	}

	return rni.HasNext, rni.Items, nil
}

func (s *Session) NewCategoryItems(rootCategoryID int) (hasNext bool, rootCategoryName string, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/new_items/%d.json", rootCategoryID))
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}

	res, err := s.Do(req)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: "+msg, rootCategoryID))
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: JSONデコードに失敗しました", rootCategoryID))
	}

	return rni.HasNext, rni.RootCategoryName, rni.Items, nil
}

func (s *Session) NewCategoryItemsWithItemIDAndCreatedAt(rootCategoryID int, itemID, createdAt int64) (hasNext bool, rootCategoryName string, items []ItemSimple, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, fmt.Sprintf("/new_items/%d.json", rootCategoryID), q)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}

	res, err := s.Do(req)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: "+msg, rootCategoryID))
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, "", nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /new_items/%d.json: JSONデコードに失敗しました", rootCategoryID))
	}

	return rni.HasNext, rni.RootCategoryName, rni.Items, nil
}

func (s *Session) UsersTransactions() (hasNext bool, items []ItemDetail, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/users/transactions.json")
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /users/transactions.json リクエストに失敗しました")
	}

	res, err := s.Do(req)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /users/transactions.json リクエストに失敗しました")
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /users/transactions.json "+msg)
	}

	rt := resTransactions{}
	err = json.NewDecoder(res.Body).Decode(&rt)
	if err != nil {
		return false, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), "GET /users/transactions.json JSONデコードに失敗しました")
	}

	return rt.HasNext, rt.Items, nil
}

func (s *Session) UserItems(userID int64) (hasNext bool, user *UserSimple, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/users/%d.json", userID))
	if err != nil {
		return false, nil, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /users/%d.json: リクエストに失敗しました", userID))
	}

	res, err := s.Do(req)
	if err != nil {
		return false, nil, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /users/%d.json: リクエストに失敗しました", userID))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /users/%d.json: "+msg, userID))
	}

	rui := resUserItems{}
	err = json.NewDecoder(res.Body).Decode(&rui)
	if err != nil {
		return false, nil, nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /users/%d.json: JSONデコードに失敗しました", userID))
	}

	return rui.HasNext, rui.User, rui.Items, nil
}

func (s *Session) Item(itemID int64) (item *ItemDetail, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/items/%d.json", itemID))
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /items/%d.json: リクエストに失敗しました", itemID))
	}

	res, err := s.Do(req)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /items/%d.json: リクエストに失敗しました", itemID))
	}
	defer res.Body.Close()

	msg, err := checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /items/%d.json: "+msg, itemID))
	}

	err = json.NewDecoder(res.Body).Decode(&item)
	if err != nil {
		return nil, fails.NewError(xerrors.Errorf("error in session: %v", err), fmt.Sprintf("GET /items/%d.json: JSONデコードに失敗しました", itemID))
	}

	return item, nil
}
