package main

import (
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
)

var (
	templates *template.Template
	dbx       *sqlx.DB
	store     sessions.Store
)

type User struct {
	ID             int64     `db:"id"`
	AccountName    string    `db:"account_name"`
	HashedPassword []byte    `db:"hashed_password"`
	CreatedAt      time.Time `db:"created_at"`
}

func init() {
	templates = template.Must(template.ParseFiles("templates/register.html", "templates/login.html"))
	store = sessions.NewCookieStore([]byte("abc"))
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

	mux.HandleFunc(pat.Get("/login"), getLogin)
	mux.HandleFunc(pat.Post("/login"), postLogin)
	mux.HandleFunc(pat.Get("/register"), getRegister)
	mux.HandleFunc(pat.Post("/register"), postRegister)

	http.ListenAndServe("localhost:8000", mux)
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", struct{}{})
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

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	err = bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	session, err := store.Get(r, sessionName)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "db error")
		return
	}

	session.Values["user_id"] = u.ID
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

	if accountName == "" || password == "" {
		outputErrorMsg(w, http.StatusInternalServerError, "all parameters are required")

		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)

		outputErrorMsg(w, http.StatusInternalServerError, "error")
		return
	}

	_, err = dbx.Exec("INSERT INTO `users` (`account_name`, `hashed_password`) VALUES (?, ?)", accountName, hashedPassword)
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
