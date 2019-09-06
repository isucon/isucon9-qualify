package session

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/isucon/isucon9-qualify/bench/asset"
	"github.com/isucon/isucon9-qualify/bench/fails"
	"github.com/morikuni/failure"
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
	ImageName   string    `json:"image_name" db:"image_name"`
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
	ImageURL   string      `json:"image_url"`
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
	ImageURL                  string      `json:"image_url"`
	CategoryID                int         `json:"category_id"`
	Category                  *Category   `json:"category"`
	TransactionEvidenceID     int64       `json:"transaction_evidence_id,omitempty"`
	TransactionEvidenceStatus string      `json:"transaction_evidence_status,omitempty"`
	ShippingStatus            string      `json:"shipping_status,omitempty"`
	CreatedAt                 int64       `json:"created_at"`
}

type TransactionEvidence struct {
	ID                 int64  `json:"id" db:"id"`
	SellerID           int64  `json:"seller_id" db:"seller_id"`
	BuyerID            int64  `json:"buyer_id" db:"buyer_id"`
	Status             string `json:"status" db:"status"`
	ItemID             int64  `json:"item_id" db:"item_id"`
	ItemName           string `json:"item_name" db:"item_name"`
	ItemPrice          int    `json:"item_price" db:"item_price"`
	ItemDescription    string `json:"item_description" db:"item_description"`
	ItemCategoryID     int    `json:"item_category_id" db:"item_category_id"`
	ItemRootCategoryID int    `json:"item_root_category_id" db:"item_root_category_id"`
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

type reqInitialize struct {
	PaymentServiceURL  string `json:"payment_service_url"`
	ShipmentServiceURL string `json:"shipment_service_url"`
}

type resInitialize struct {
	Campaign int    `json:"campaign"`
	Language string `json:"language"`
}

type resSetting struct {
	CSRFToken         string     `json:"csrf_token"`
	PaymentServiceURL string     `json:"payment_service_url"`
	User              *User      `json:"user,omitempty"`
	Categories        []Category `json:"categories"`
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

type reqShip struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
}

type resShip struct {
	Path      string `json:"path"`
	ReserveID string `json:"reserve_id"`
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

func (s *Session) Initialize(ctx context.Context, paymentServiceURL, shipmentServiceURL string) (int, string, error) {
	b, _ := json.Marshal(reqInitialize{
		PaymentServiceURL:  paymentServiceURL,
		ShipmentServiceURL: shipmentServiceURL,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/initialize", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, "", failure.Wrap(err, failure.Message("POST /initialize: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, "", failure.Wrap(err, failure.Message("POST /initialize: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, "", err
	}

	ri := resInitialize{}
	err = json.NewDecoder(res.Body).Decode(&ri)
	if err != nil {
		return 0, "", failure.Wrap(err, failure.Message("POST /initialize: JSONデコードに失敗しました"))
	}

	return ri.Campaign, ri.Language, nil
}

func (s *Session) Login(ctx context.Context, accountName, password string) (*asset.AppUser, error) {
	b, _ := json.Marshal(reqLogin{
		AccountName: accountName,
		Password:    password,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("POST /login: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	u := &asset.AppUser{}
	err = json.NewDecoder(res.Body).Decode(u)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("POST /login: JSONデコードに失敗しました"))
	}

	return u, nil
}

func (s *Session) SetSettings(ctx context.Context) error {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/settings")
	if err != nil {
		return failure.Wrap(err, failure.Message("GET /settings: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Message("GET /settings: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return err
	}

	rs := &resSetting{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return failure.Wrap(err, failure.Message("GET /settings: JSONデコードに失敗しました"))
	}

	if rs.CSRFToken == "" {
		return failure.New(fails.ErrApplication, failure.Message("GET /settings: csrf tokenが空です"))
	}

	if rs.User == nil || rs.User.ID == 0 {
		return failure.New(fails.ErrApplication, failure.Message("GET /settings: userが空です"))
	}

	s.UserID = rs.User.ID
	s.csrfToken = rs.CSRFToken
	return nil
}

func (s *Session) Sell(ctx context.Context, fileName, name string, price int, description string, categoryID int) (int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: 画像のOpenに失敗しました"))
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "upload.jpg")
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	writer.WriteField("csrf_token", s.csrfToken)
	writer.WriteField("name", name)
	writer.WriteField("description", description)
	writer.WriteField("price", strconv.Itoa(price))
	writer.WriteField("category_id", strconv.Itoa(categoryID))

	contentType := writer.FormDataContentType()

	err = writer.Close()
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/sell", contentType, body)
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return 0, err
	}

	rs := &resSell{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return 0, failure.Wrap(err, failure.Message("POST /sell: JSONデコードに失敗しました"))
	}

	return rs.ID, nil
}

func (s *Session) Buy(ctx context.Context, itemID int64, token string) (int64, error) {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return 0, err
	}

	rb := &resBuy{}
	err = json.NewDecoder(res.Body).Decode(rb)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return rb.TransactionEvidenceID, nil
}

// 人気者出品用。成功するかもしれないし、失敗するかもしれない。
// この中では異質だが正常系ではあるのでここで定義する
func (s *Session) BuyWithMayFail(ctx context.Context, itemID int64, token string) (int64, error) {
	b, _ := json.Marshal(reqBuy{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		Token:     token,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/buy", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		re := resErr{}
		err = json.NewDecoder(res.Body).Decode(&re)
		if err != nil {
			return 0, failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
		}

		expectedMsg := "item is not for sale"

		if re.Error != expectedMsg {
			return 0, failure.Wrap(err, failure.Messagef("POST /buy: exected error message: %s; actual: %s (item_id: %d)", expectedMsg, re.Error, itemID))
		}

		// イレギュラーだが、エラーがないのに0が返っていたら正常に買えなかったという扱いにする
		return 0, nil
	}

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return 0, err
	}

	rb := &resBuy{}
	err = json.NewDecoder(res.Body).Decode(rb)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /buy: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return rb.TransactionEvidenceID, nil
}

func (s *Session) Ship(ctx context.Context, itemID int64) (reserveID, apath string, err error) {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", "", failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return "", "", failure.Wrap(err, failure.Messagef("POST /ship: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return "", "", err
	}

	rs := &resShip{}
	err = json.NewDecoder(res.Body).Decode(rs)
	if err != nil {
		return "", "", failure.Wrap(err, failure.Messagef("POST /ship: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	if len(rs.Path) == 0 {
		return "", "", failure.New(fails.ErrApplication, failure.Messagef("POST /ship: pathが空です (item_id: %d)", itemID))
	}

	if len(rs.ReserveID) == 0 {
		return "", "", failure.New(fails.ErrApplication, failure.Messagef("POST /ship: reserve_idが空です (item_id: %d)", itemID))
	}

	return rs.ReserveID, rs.Path, nil
}

func (s *Session) ShipDone(ctx context.Context, itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/ship_done", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /ship_done: bodyの読み込みに失敗しました (item_id: %d)", itemID))
	}

	return nil
}

func (s *Session) Complete(ctx context.Context, itemID int64) error {
	b, _ := json.Marshal(reqShip{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/complete", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /complete: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /complete: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return err
	}

	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return failure.Wrap(err, failure.Messagef("POST /complete: bodyの読み込みに失敗しました (item_id: %d)", itemID))
	}

	return nil
}

func (s *Session) DownloadQRURL(ctx context.Context, apath string) (md5Str string, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, res.Body)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: bodyの読み込みに失敗しました", apath))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *Session) DownloadItemImageURL(ctx context.Context, apath string) (md5Str string, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, res.Body)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: bodyの読み込みに失敗しました", apath))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *Session) DownloadStaticURL(ctx context.Context, apath string) (md5Str string, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, apath)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: リクエストに失敗しました", apath))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = io.Copy(h, res.Body)
	if err != nil {
		return "", failure.Wrap(err, failure.Messagef("GET %s: bodyの読み込みに失敗しました", apath))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func (s *Session) Bump(ctx context.Context, itemID int64) (int64, error) {
	b, _ := json.Marshal(reqBump{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/bump", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /bump: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /bump: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return 0, err
	}

	rie := &resItemEdit{}
	err = json.NewDecoder(res.Body).Decode(rie)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /bump: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return rie.ItemCreatedAt, nil
}

func (s *Session) ItemEdit(ctx context.Context, itemID int64, price int) (int, error) {
	b, _ := json.Marshal(reqItemEdit{
		CSRFToken: s.csrfToken,
		ItemID:    itemID,
		ItemPrice: price,
	})
	req, err := s.newPostRequest(ShareTargetURLs.AppURL, "/items/edit", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /items/edit: リクエストに失敗しました (item_id: %d)", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /items/edit: リクエストに失敗しました (item_id: %d)", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCodeWithMsg(res, http.StatusOK, fmt.Sprintf("(item_id: %d)", itemID))
	if err != nil {
		return 0, err
	}

	rie := &resItemEdit{}
	err = json.NewDecoder(res.Body).Decode(rie)
	if err != nil {
		return 0, failure.Wrap(err, failure.Messagef("POST /items/edit: JSONデコードに失敗しました (item_id: %d)", itemID))
	}

	return rie.ItemPrice, nil
}

func (s *Session) NewItems(ctx context.Context) (hasNext bool, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/new_items.json")
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, err
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: JSONデコードに失敗しました"))
	}

	return rni.HasNext, rni.Items, nil
}

func (s *Session) NewItemsWithItemIDAndCreatedAt(ctx context.Context, itemID, createdAt int64) (hasNext bool, items []ItemSimple, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, "/new_items.json", q)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, err
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Message("GET /new_items.json: JSONデコードに失敗しました"))
	}

	return rni.HasNext, rni.Items, nil
}

func (s *Session) NewCategoryItems(ctx context.Context, rootCategoryID int) (hasNext bool, rootCategoryName string, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/new_items/%d.json", rootCategoryID))
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, "", nil, err
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: JSONデコードに失敗しました", rootCategoryID))
	}

	return rni.HasNext, rni.RootCategoryName, rni.Items, nil
}

func (s *Session) NewCategoryItemsWithItemIDAndCreatedAt(ctx context.Context, rootCategoryID int, itemID, createdAt int64) (hasNext bool, rootCategoryName string, items []ItemSimple, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, fmt.Sprintf("/new_items/%d.json", rootCategoryID), q)
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: リクエストに失敗しました", rootCategoryID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, "", nil, err
	}

	rni := resNewItems{}
	err = json.NewDecoder(res.Body).Decode(&rni)
	if err != nil {
		return false, "", nil, failure.Wrap(err, failure.Messagef("GET /new_items/%d.json: JSONデコードに失敗しました", rootCategoryID))
	}

	return rni.HasNext, rni.RootCategoryName, rni.Items, nil
}

func (s *Session) UsersTransactions(ctx context.Context) (hasNext bool, items []ItemDetail, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/users/transactions.json")
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json リクエストに失敗しました (user_id: %d)", s.UserID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json リクエストに失敗しました (user_id: %d)", s.UserID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, err
	}

	rt := resTransactions{}
	err = json.NewDecoder(res.Body).Decode(&rt)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json JSONデコードに失敗しました (user_id: %d)", s.UserID))
	}

	return rt.HasNext, rt.Items, nil
}

func (s *Session) UsersTransactionsWithItemIDAndCreatedAt(ctx context.Context, itemID, createdAt int64) (hasNext bool, items []ItemDetail, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, "/users/transactions.json", q)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json リクエストに失敗しました (user_id: %d)", s.UserID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json リクエストに失敗しました (user_id: %d)", s.UserID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, err
	}

	rt := resTransactions{}
	err = json.NewDecoder(res.Body).Decode(&rt)
	if err != nil {
		return false, nil, failure.Wrap(err, failure.Messagef("GET /users/transactions.json JSONデコードに失敗しました (user_id: %d)", s.UserID))
	}

	return rt.HasNext, rt.Items, nil
}

func (s *Session) UserItems(ctx context.Context, userID int64) (hasNext bool, user *UserSimple, items []ItemSimple, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/users/%d.json", userID))
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: リクエストに失敗しました", userID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: リクエストに失敗しました", userID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, nil, err
	}

	rui := resUserItems{}
	err = json.NewDecoder(res.Body).Decode(&rui)
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: JSONデコードに失敗しました", userID))
	}

	return rui.HasNext, rui.User, rui.Items, nil
}

func (s *Session) UserItemsWithItemIDAndCreatedAt(ctx context.Context, userID, itemID, createdAt int64) (hasNext bool, user *UserSimple, items []ItemSimple, err error) {
	q := url.Values{}
	q.Set("item_id", strconv.FormatInt(itemID, 10))
	q.Set("created_at", strconv.FormatInt(createdAt, 10))

	req, err := s.newGetRequestWithQuery(ShareTargetURLs.AppURL, fmt.Sprintf("/users/%d.json", userID), q)
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: リクエストに失敗しました", userID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: リクエストに失敗しました", userID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return false, nil, nil, err
	}

	rui := resUserItems{}
	err = json.NewDecoder(res.Body).Decode(&rui)
	if err != nil {
		return false, nil, nil, failure.Wrap(err, failure.Messagef("GET /users/%d.json: JSONデコードに失敗しました", userID))
	}

	return rui.HasNext, rui.User, rui.Items, nil
}

func (s *Session) Item(ctx context.Context, itemID int64) (item ItemDetail, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, fmt.Sprintf("/items/%d.json", itemID))
	if err != nil {
		return ItemDetail{}, failure.Wrap(err, failure.Messagef("GET /items/%d.json: リクエストに失敗しました", itemID))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return ItemDetail{}, failure.Wrap(err, failure.Messagef("GET /items/%d.json: リクエストに失敗しました", itemID))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return ItemDetail{}, err
	}

	err = json.NewDecoder(res.Body).Decode(&item)
	if err != nil {
		return ItemDetail{}, failure.Wrap(err, failure.Messagef("GET /items/%d.json: JSONデコードに失敗しました", itemID))
	}

	return item, nil
}

func (s *Session) Reports(ctx context.Context) (transactionEvidences []TransactionEvidence, err error) {
	req, err := s.newGetRequest(ShareTargetURLs.AppURL, "/reports.json")
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /reports.json: リクエストに失敗しました"))
	}

	req = req.WithContext(ctx)

	res, err := s.Do(req)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /reports.json: リクエストに失敗しました"))
	}
	defer res.Body.Close()

	err = checkStatusCode(res, http.StatusOK)
	if err != nil {
		return nil, err
	}

	transactionEvidences = make([]TransactionEvidence, 0, 100)

	err = json.NewDecoder(res.Body).Decode(&transactionEvidences)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("GET /reports.json: JSONデコードに失敗しました"))
	}

	return transactionEvidences, nil
}
