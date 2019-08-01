package main

import (
	crand "crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/isucon/isucon9-qualify/webapp/go/api"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	goji "goji.io"
	"goji.io/pat"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionName = "session_isucari"

	ItemStatusOnSale  = "on_sale"
	ItemStatusTrading = "trading"
	ItemStatusSoldOut = "sold_out"
	ItemStatusStop    = "stop"
	ItemStatusCancel  = "cancel"

	PaymentServiceIsucariAPIKey = "a15400e46c83635eb181-946abb51ff26a868317c"
	PaymentServiceIsucariShopID = "11"

	TransactionEvidenceStatusWaitShipping = "wait_shipping"
	TransactionEvidenceStatusWaitDone     = "wait_done"
	TransactionEvidenceStatusDone         = "done"

	ShippingsStatusInitial    = "initial"
	ShippingsStatusWaitPickup = "wait_pickup"
	ShippingsStatusShipping   = "shipping"
	ShippingsStatusDone       = "done"
)

var (
	templates *template.Template
	dbx       *sqlx.DB
	store     sessions.Store
)

type User struct {
	ID             int64     `json:"id" db:"id"`
	AccountName    string    `json:"account_name" db:"account_name"`
	HashedPassword []byte    `json:"-" db:"hashed_password"`
	Address        string    `json:"address,omitempty" db:"address"`
	CreatedAt      time.Time `json:"-" db:"created_at"`
}

type Item struct {
	ID          int64     `json:"id" db:"id"`
	SellerID    int64     `json:"seller_id" db:"seller_id"`
	BuyerID     int64     `json:"buyer_id" db:"buyer_id"`
	Status      string    `json:"status" db:"status"`
	Name        string    `json:"name" db:"name"`
	Price       int       `json:"price" db:"price"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type TransactionEvidence struct {
	ID              int64     `json:"id" db:"id"`
	SellerID        int64     `json:"seller_id" db:"seller_id"`
	BuyerID         int64     `json:"buyer_id" db:"buyer_id"`
	Status          string    `json:"status" db:"status"`
	ItemID          int64     `json:"item_id" db:"item_id"`
	ItemName        string    `json:"item_name" db:"item_name"`
	ItemPrice       int       `json:"item_price" db:"item_price"`
	ItemDescription string    `json:"item_description" db:"item_description"`
	CreatedAt       time.Time `json:"-" db:"created_at"`
	UpdatedAt       time.Time `json:"-" db:"updated_at"`
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
	ImgName               string    `json:"img_name" db:"img_name"`
	CreatedAt             time.Time `json:"-" db:"created_at"`
	UpdatedAt             time.Time `json:"-" db:"updated_at"`
}

type reqBuy struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	Token     string `json:"token"`
}

type resSell struct {
	ID int64 `json:"id"`
}

type resPostShip struct {
	URL string `json:"url"`
}

func init() {
	templates = template.Must(template.ParseFiles(
		"templates/register.html",
		"templates/login.html",
		"templates/item_edit.html",
		"templates/sell.html",
		"templates/buy.html",
		"templates/ship.html",
		"templates/ship_done.html",
		"templates/complete.html",
	))
	store = sessions.NewCookieStore([]byte("abc"))

	log.SetFlags(log.Lshortfile)
}

func main() {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("failed to read DB port number from an environment variable MYSQL_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "isucari"
	}
	dbname := os.Getenv("MYSQL_DBNAME")
	if dbname == "" {
		dbname = "isucari"
	}
	password := os.Getenv("MYSQL_PASS")
	if password == "" {
		password = "isucari"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	dbx, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to DB: %s.", err.Error())
	}
	defer dbx.Close()

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/items/:item_id.json"), getItem)
	mux.HandleFunc(pat.Get("/items/:item_id/edit"), getItemEdit)
	mux.HandleFunc(pat.Post("/items/edit"), postItemEdit)
	mux.HandleFunc(pat.Get("/buy/:item_id"), getBuyItem)
	mux.HandleFunc(pat.Post("/buy"), postBuy)
	mux.HandleFunc(pat.Get("/sell"), getSell)
	mux.HandleFunc(pat.Post("/sell"), postSell)
	mux.HandleFunc(pat.Get("/ship/:item_id"), getShip)
	mux.HandleFunc(pat.Post("/ship"), postShip)
	mux.HandleFunc(pat.Get("/ship_done/:item_id"), getShipDone)
	mux.HandleFunc(pat.Post("/ship_done"), postShipDone)
	mux.HandleFunc(pat.Get("/complete/:item_id"), getComplete)
	mux.HandleFunc(pat.Post("/complete"), postComplete)
	mux.HandleFunc(pat.Get("/login"), getLogin)
	mux.HandleFunc(pat.Post("/login"), postLogin)
	mux.HandleFunc(pat.Get("/register"), getRegister)
	mux.HandleFunc(pat.Post("/register"), postRegister)
	mux.Handle(pat.Get("/*"), http.FileServer(http.Dir("../public")))

	http.ListenAndServe("localhost:8000", mux)
}

func getSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, sessionName)

	return session
}

func getCSRFToken(r *http.Request) string {
	session := getSession(r)

	csrfToken, ok := session.Values["csrf_token"]
	if !ok {
		return ""
	}

	return csrfToken.(string)
}

func getUserID(r *http.Request) int64 {
	session := getSession(r)

	userID, ok := session.Values["user_id"]
	if !ok {
		return 0
	}

	return userID.(int64)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	itemID := pat.Param(r, "item_id")

	item := Item{}
	err := dbx.Get(&item, "SELECT * FROM `items` WHERE `id` = ?", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	json.NewEncoder(w).Encode(item)
}

func getItemEdit(w http.ResponseWriter, r *http.Request) {
	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusBadRequest, "id error")
		return
	}

	csrfToken := getCSRFToken(r)

	templates.ExecuteTemplate(w, "item_edit.html", struct {
		CSRFToken string
		ItemID    int64
	}{csrfToken, itemID})
}

func postItemEdit(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	priceStr := r.FormValue("item_price")
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	if csrfToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	userID := getUserID(r)

	targetItem := Item{}
	err = dbx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ?", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	if targetItem.SellerID != userID {
		outputErrorMsg(w, http.StatusForbidden, "自分の商品以外は編集できません")
		return
	}

	if targetItem.Status != ItemStatusOnSale {
		outputErrorMsg(w, http.StatusForbidden, "販売中の商品以外編集できません")
		return
	}

	tx := dbx.MustBegin()
	err = tx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", itemID)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `items` SET `price` = ?, `updated_at` = ? WHERE `id` = ?",
		price,
		time.Now(),
		itemID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	tx.Commit()
}

func getBuyItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusBadRequest, "id error")
		return
	}

	csrfToken := getCSRFToken(r)

	templates.ExecuteTemplate(w, "buy.html", struct {
		CSRFToken string
		ItemID    int64

		PaymentServiceShopID string
	}{csrfToken, itemID, PaymentServiceIsucariShopID})
}

func postBuy(w http.ResponseWriter, r *http.Request) {
	rb := reqBuy{}

	err := json.NewDecoder(r.Body).Decode(&rb)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "json decode error")
		return
	}

	if rb.CSRFToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	buyerID := getUserID(r)

	targetItem := Item{}
	err = dbx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ?", rb.ItemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	if targetItem.Status != ItemStatusOnSale {
		outputErrorMsg(w, http.StatusForbidden, "item is not for sale")
		return
	}

	if targetItem.SellerID == buyerID {
		outputErrorMsg(w, http.StatusForbidden, "自分の商品は買えません")
		return
	}

	tx := dbx.MustBegin()
	err = tx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", rb.ItemID)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	buyer := User{}
	seller := User{}
	err = dbx.Get(&buyer, "SELECT * FROM `users` WHERE `id` = ?", buyerID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "user not found")
		tx.Rollback()
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	err = dbx.Get(&seller, "SELECT * FROM `users` WHERE `id` = ?", targetItem.SellerID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "user not found")
		tx.Rollback()
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	result, err := tx.Exec("INSERT INTO `transaction_evidences` (`seller_id`, `buyer_id`, `status`, `item_id`, `item_name`, `item_price`, `item_description`) VALUES (?, ?, ?, ?, ?, ?, ?)",
		targetItem.SellerID,
		buyerID,
		TransactionEvidenceStatusWaitShipping,
		targetItem.ID,
		targetItem.Name,
		targetItem.Price,
		targetItem.Description,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	transactionEvidenceID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `items` SET `buyer_id` = ?, `status` = ?, `updated_at` = ? WHERE `id` = ?",
		buyerID,
		ItemStatusTrading,
		time.Now(),
		targetItem.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	scr, err := api.ShipmentCreate("http://localhost:7000", &api.ShipmentCreateReq{
		ToAddress:   buyer.Address,
		ToName:      buyer.AccountName,
		FromAddress: seller.Address,
		FromName:    seller.AccountName,
	})
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "failed to request to shipment service")
		tx.Rollback()

		return
	}

	pstr, err := api.PaymentToken("http://localhost:5555", &api.PaymentServiceTokenReq{
		Token:  rb.Token,
		APIKey: PaymentServiceIsucariAPIKey,
		Price:  targetItem.Price,
	})
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "payment service is failed")
		tx.Rollback()
		return
	}

	if pstr.Status == "invalid" {
		outputErrorMsg(w, http.StatusBadRequest, "カード情報に誤りがあります")
		tx.Rollback()
		return
	}

	if pstr.Status == "fail" {
		outputErrorMsg(w, http.StatusBadRequest, "カードの残高が足りません")
		tx.Rollback()
		return
	}

	if pstr.Status != "ok" {
		outputErrorMsg(w, http.StatusBadRequest, "想定外のエラー")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT INTO `shippings` (`transaction_evidence_id`, `status`, `item_name`, `item_id`, `reserve_id`, `reserve_time`, `to_address`, `to_name`, `from_address`, `from_name`, `img_name`) VALUES (?,?,?,?,?,?,?,?,?,?,?)",
		transactionEvidenceID,
		ShippingsStatusInitial,
		targetItem.Name,
		targetItem.ID,
		scr.ReserveID,
		scr.ReserveTime,
		buyer.Address,
		buyer.AccountName,
		seller.Address,
		seller.AccountName,
		"",
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	tx.Commit()
}

func getShip(w http.ResponseWriter, r *http.Request) {
	csrfToken := getCSRFToken(r)

	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	templates.ExecuteTemplate(w, "ship.html", struct {
		CSRFToken string
		ItemID    int64
	}{csrfToken, itemID})
}

func postShip(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	if csrfToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	userID := getUserID(r)
	transactionEvidence := TransactionEvidence{}
	err = dbx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "transaction_evidences not found")
		return
	}
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")

		return
	}

	if transactionEvidence.SellerID != userID {
		outputErrorMsg(w, http.StatusForbidden, "権限がありません")
		return
	}

	if transactionEvidence.Status != TransactionEvidenceStatusWaitShipping {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		return
	}

	tx := dbx.MustBegin()
	err = tx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ? FOR UPDATE", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "db not found")
		tx.Rollback()
		return
	}
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}
	if transactionEvidence.Status != TransactionEvidenceStatusWaitShipping {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		tx.Rollback()
		return
	}
	item := Item{}
	err = tx.Get(&item, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "item not found")
		tx.Rollback()
		return
	}
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	shipping := Shipping{}
	err = tx.Get(&shipping, "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE", transactionEvidence.ID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusNotFound, "shippings not found")
		tx.Rollback()
		return
	}
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	img, err := api.ShipmentRequest("http://localhost:7000", &api.ShipmentRequestReq{
		ReserveID: shipping.ReserveID,
	})
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "failed to request to shipment service")
		tx.Rollback()

		return
	}

	imgName := secureRandomStr(16)

	err = ioutil.WriteFile(fmt.Sprintf("../public/upload/%s.png", imgName), img, 0644)
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "failed to save the file")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `shippings` SET `status` = ?, `img_name` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?",
		ShippingsStatusWaitPickup,
		imgName,
		time.Now(),
		transactionEvidence.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	tx.Commit()

	rps := resPostShip{
		URL: fmt.Sprintf("http://%s/upload/%s.png", r.Host, imgName),
	}
	json.NewEncoder(w).Encode(rps)
}

func getShipDone(w http.ResponseWriter, r *http.Request) {
	csrfToken := getCSRFToken(r)

	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	templates.ExecuteTemplate(w, "ship_done.html", struct {
		CSRFToken string
		ItemID    int64
	}{csrfToken, itemID})
}

func postShipDone(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	if csrfToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	userID := getUserID(r)
	transactionEvidence := TransactionEvidence{}
	err = dbx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", itemID)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusBadRequest, "transaction_evidence not found")
		return
	}
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")

		return
	}

	if transactionEvidence.SellerID != userID {
		outputErrorMsg(w, http.StatusForbidden, "権限がありません")
		return
	}

	if transactionEvidence.Status != TransactionEvidenceStatusWaitShipping {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		return
	}

	tx := dbx.MustBegin()
	tx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ? FOR UPDATE", itemID)
	if transactionEvidence.Status != TransactionEvidenceStatusWaitShipping {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		tx.Rollback()
		return
	}
	item := Item{}
	tx.Get(&item, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", itemID)

	shipping := Shipping{}
	err = tx.Get(&shipping, "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE", transactionEvidence.ID)
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	ssr, err := api.ShipmentStatus("http://localhost:7000", &api.ShipmentStatusReq{
		ReserveID: shipping.ReserveID,
	})
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "failed to request to shipment service")
		tx.Rollback()

		return
	}

	if !(ssr.Status == ShippingsStatusShipping || ssr.Status == ShippingsStatusDone) {
		outputErrorMsg(w, http.StatusInternalServerError, "shipment service側で配送中か配送完了になっていません")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?",
		ssr.Status,
		time.Now(),
		transactionEvidence.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?",
		TransactionEvidenceStatusWaitDone,
		time.Now(),
		transactionEvidence.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	tx.Commit()
}

func getComplete(w http.ResponseWriter, r *http.Request) {
	csrfToken := getCSRFToken(r)

	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	templates.ExecuteTemplate(w, "complete.html", struct {
		CSRFToken string
		ItemID    int64
	}{csrfToken, itemID})
}

func postComplete(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		outputErrorMsg(w, http.StatusBadRequest, "invalid syntax")
		return
	}

	if csrfToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	userID := getUserID(r)
	transactionEvidence := TransactionEvidence{}
	err = dbx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ?", itemID)
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")

		return
	}

	if transactionEvidence.BuyerID != userID {
		outputErrorMsg(w, http.StatusForbidden, "権限がありません")
		return
	}

	if transactionEvidence.Status != TransactionEvidenceStatusWaitDone {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		return
	}

	tx := dbx.MustBegin()
	tx.Get(&transactionEvidence, "SELECT * FROM `transaction_evidences` WHERE `item_id` = ? FOR UPDATE", itemID)
	if transactionEvidence.Status != TransactionEvidenceStatusWaitDone {
		outputErrorMsg(w, http.StatusForbidden, "準備ができていません")
		tx.Rollback()
		return
	}
	item := Item{}
	tx.Get(&item, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", itemID)

	shipping := Shipping{}
	err = tx.Get(&shipping, "SELECT * FROM `shippings` WHERE `transaction_evidence_id` = ? FOR UPDATE", transactionEvidence.ID)
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	ssr, err := api.ShipmentStatus("http://localhost:7000", &api.ShipmentStatusReq{
		ReserveID: shipping.ReserveID,
	})
	if err != nil {
		log.Println(err)
		outputErrorMsg(w, http.StatusInternalServerError, "failed to request to shipment service")
		tx.Rollback()

		return
	}

	if !(ssr.Status == ShippingsStatusDone) {
		outputErrorMsg(w, http.StatusInternalServerError, "shipment service側で配送完了になっていません")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `items` SET `status` = ?, `updated_at` = ? WHERE `id` = ?",
		ItemStatusSoldOut,
		time.Now(),
		itemID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `shippings` SET `status` = ?, `updated_at` = ? WHERE `transaction_evidence_id` = ?",
		ShippingsStatusDone,
		time.Now(),
		transactionEvidence.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("UPDATE `transaction_evidences` SET `status` = ?, `updated_at` = ? WHERE `id` = ?",
		TransactionEvidenceStatusDone,
		time.Now(),
		transactionEvidence.ID,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		tx.Rollback()
		return
	}

	tx.Commit()
}

func getSell(w http.ResponseWriter, r *http.Request) {
	csrfToken := getCSRFToken(r)

	templates.ExecuteTemplate(w, "sell.html", struct {
		CSRFToken string
	}{csrfToken})
}

func postSell(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	name := r.FormValue("name")
	price := r.FormValue("price")
	description := r.FormValue("description")

	if csrfToken != getCSRFToken(r) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	sellerID := getUserID(r)

	result, err := dbx.Exec("INSERT INTO `items` (`seller_id`, `status`, `name`, `price`, `description`) VALUES (?, ?, ?, ?, ?)",
		sellerID,
		ItemStatusOnSale,
		name,
		price,
		description,
	)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	itemID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	json.NewEncoder(w).Encode(resSell{ID: itemID})
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", struct{}{})
}

func secureRandomStr(b int) string {
	k := make([]byte, b)
	if _, err := crand.Read(k); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", k)
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	accountName := r.FormValue("account_name")
	password := r.FormValue("password")

	if accountName == "" || password == "" {
		outputErrorMsg(w, http.StatusBadRequest, "all parameters are required")

		return
	}

	u := User{}
	err := dbx.Get(&u, "SELECT * FROM `users` WHERE `account_name` = ?", accountName)
	if err == sql.ErrNoRows {
		outputErrorMsg(w, http.StatusUnauthorized, "user not found")
		return
	}
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	err = bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "crypt error")
		return
	}

	session := getSession(r)

	session.Values["user_id"] = u.ID
	session.Values["csrf_token"] = secureRandomStr(20)
	if err = session.Save(r, w); err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", struct{}{})
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	accountName := r.FormValue("account_name")
	password := r.FormValue("password")
	address := r.FormValue("address")

	if accountName == "" || password == "" || address == "" {
		outputErrorMsg(w, http.StatusInternalServerError, "all parameters are required")

		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "error")
		return
	}

	_, err = dbx.Exec("INSERT INTO `users` (`account_name`, `hashed_password`, `address`) VALUES (?, ?, ?)", accountName, hashedPassword, address)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func outputErrorMsg(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
}
