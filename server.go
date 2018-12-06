package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"net/http"
	"database/sql"
	"html/template"
	_ "github.com/lib/pq"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var tpl *template.Template
var router *mux.Router
var db *sql.DB
var context Business
var err error
var username = os.Getenv("DB_USERNAME")
var password = os.Getenv("DB_PASSWORD")
var database = os.Getenv("DB_NAME")
var users_table = os.Getenv("USERS_TABLE")
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

func init() {

	// Quick check on environment variables.
	fmt.Println("USERNAME: ", username)
	fmt.Println("PASSWORD: ", password)
	fmt.Println("SESSION KEY: ", os.Getenv("SESSION_KEY"))

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i+1
		},
		"format": func(i time.Time) string {
			return i.Format("2006-01-02")
		},
		"remain": func(s string) string {
			return strings.Join(strings.Split(strings.Split(s, ".")[0], ":")[:2], ":")
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"voted": func(b bool) string {
			if b {
				return "Voted"
			} else {
				return "Did not Vote."
			}
		},
	}
	
	tpl = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
	router = mux.NewRouter()
	context = Business{
		Logo: "Wombol.com",
		Slogan: "Social Commerce",
		DbName: database,
		DbUser: username,
		DbPassword: password,
		UserTable: users_table,
		Message: "",
		Data: "",
	}

	db, err = sql.Open("postgres", context.dbInfo())
	fmt.Println(err)
	fmt.Println("DB ERROR: ", db)

	err = db.Ping()
	fmt.Println("PING ERROR:", err)
	if err != nil {
		fmt.Println("Failed to connect to the database.")
		fmt.Println(err)
	}

}

func main() {
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/", index_handler).Methods("GET")
	router.HandleFunc("/", auth_handler).Methods("POST")

	//router.HandleFunc("/contact/", contact_handler).Methods("GET")
	//router.HandleFunc("/message/", message_handler).Methods("POST")

	port := ":8000"
	fmt.Println("Listening to port " + port)
	http.ListenAndServe(port, router)
	defer db.Close()
}
