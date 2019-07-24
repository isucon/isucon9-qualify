package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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

	TransactionEvidenceStatusWaitPayment  = "wait_payment"
	TransactionEvidenceStatusWaitShipping = "wait_shipping"
	TransactionEvidenceStatusWaitDone     = "wait_done"
	TransactionEvidenceStatusDone         = "done"
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

type reqBuy struct {
	CSRFToken string `json:"csrf_token"`
	ItemID    int64  `json:"item_id"`
	Token     string `json:"token"`
}

func init() {
	templates = template.Must(template.ParseFiles(
		"templates/register.html",
		"templates/login.html",
		"templates/sell.html",
		"templates/buy.html",
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
	mux.HandleFunc(pat.Get("/buy/:item_id"), getBuyItem)
	mux.HandleFunc(pat.Post("/buy"), postBuy)
	mux.HandleFunc(pat.Get("/sell"), getSell)
	mux.HandleFunc(pat.Post("/sell"), postSell)
	mux.HandleFunc(pat.Get("/login"), getLogin)
	mux.HandleFunc(pat.Post("/login"), postLogin)
	mux.HandleFunc(pat.Get("/register"), getRegister)
	mux.HandleFunc(pat.Post("/register"), postRegister)
	mux.Handle(pat.Get("/*"), http.FileServer(http.Dir("../public")))

	http.ListenAndServe("localhost:8000", mux)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	itemID := pat.Param(r, "item_id")

	item := Item{}
	err := dbx.Get(&item, "SELECT * FROM `items` WHERE `id` = ?", itemID)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	b, _ := json.Marshal(item)
	w.Write(b)
}

func getBuyItem(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	itemIDStr := pat.Param(r, "item_id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "id error")
		return
	}

	csrfToken := session.Values["csrf_token"].(string)

	templates.ExecuteTemplate(w, "buy.html", struct {
		CSRFToken string
		ItemID    int64

		PaymentServiceShopID string
	}{csrfToken, itemID, PaymentServiceIsucariShopID})
}

type paymentServiceTokenReq struct {
	Token  string `json:"token"`
	APIKey string `json:"api_key"`
	Price  int    `json:"price"`
}

type paymentServiceTokenRes struct {
	Status string `json:"status"`
}

func postBuy(w http.ResponseWriter, r *http.Request) {
	rb := reqBuy{}

	err := json.NewDecoder(r.Body).Decode(&rb)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "json decode error")
		return
	}

	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	if rb.CSRFToken != session.Values["csrf_token"].(string) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	buyerID := session.Values["user_id"].(int64)

	targetItem := Item{}
	dbx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ?", rb.ItemID)
	if targetItem.ID != rb.ItemID {
		outputErrorMsg(w, http.StatusNotFound, "item not exist")
		return
	}

	if targetItem.Status != ItemStatusOnSale {
		outputErrorMsg(w, http.StatusNotFound, "item is not for sale")
		return
	}

	if targetItem.SellerID == buyerID {
		outputErrorMsg(w, http.StatusNotFound, "自分の商品は買えません")
		return
	}

	tx := dbx.MustBegin()
	err = tx.Get(&targetItem, "SELECT * FROM `items` WHERE `id` = ? FOR UPDATE", rb.ItemID)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		tx.Rollback()
		return
	}

	_, err = tx.Exec("INSERT INTO `transaction_evidences` (`seller_id`, `buyer_id`, `status`, `item_id`, `item_name`, `item_price`, `item_description`) VALUES (?, ?, ?, ?, ?, ?, ?)",
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
	_, err = tx.Exec("UPDATE `items` SET buyer_id = ?, status = ?, updated_at = ? WHERE id = ?",
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

	body := &paymentServiceTokenReq{
		Token:  rb.Token,
		APIKey: PaymentServiceIsucariAPIKey,
		Price:  100,
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post("http://localhost:5555/token", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "failed to request to payment service")
		tx.Rollback()
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("payment service's status is %d", resp.StatusCode)

		outputErrorMsg(w, http.StatusInternalServerError, "payment service is failed")
		tx.Rollback()
		return
	}

	pstr := &paymentServiceTokenRes{}
	err = json.NewDecoder(resp.Body).Decode(&pstr)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "json decode error")
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

	tx.Commit()
}

func getSell(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	csrfToken := session.Values["csrf_token"].(string)

	templates.ExecuteTemplate(w, "sell.html", struct {
		CSRFToken string
	}{csrfToken})
}

func postSell(w http.ResponseWriter, r *http.Request) {
	csrfToken := r.FormValue("csrf_token")
	name := r.FormValue("name")
	price := r.FormValue("price")
	description := r.FormValue("description")

	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	if csrfToken != session.Values["csrf_token"].(string) {
		outputErrorMsg(w, http.StatusUnprocessableEntity, "csrf token error")

		return
	}

	sellerID := session.Values["user_id"]

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

	http.Redirect(w, r, fmt.Sprintf("/items/%d.json", itemID), http.StatusFound)
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
		outputErrorMsg(w, http.StatusInternalServerError, "all parameters are required")

		return
	}

	u := User{}
	err := dbx.Get(&u, "SELECT * FROM `users` WHERE `account_name` = ?", accountName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	err = bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "crypt error")
		return
	}

	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
		return
	}

	session.Values["user_id"] = u.ID
	session.Values["csrf_token"] = secureRandomStr(20)
	if err = session.Save(r, w); err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "session error")
	}

	http.Redirect(w, r, "/", http.StatusFound)
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

	b, _ := json.Marshal(struct {
		Error string `json:"error"`
	}{Error: msg})

	w.WriteHeader(status)
	w.Write(b)
}
