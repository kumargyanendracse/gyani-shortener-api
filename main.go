package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Link struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/shorten", shorten)
	http.HandleFunc("/resolve", resolve)

	log.Println("Server running")
	http.ListenAndServe(":8080", nil)
}

func shorten(w http.ResponseWriter, r *http.Request) {
	var link Link
	json.NewDecoder(r.Body).Decode(&link)

	_, err := db.Exec("INSERT INTO links (code, url) VALUES ($1,$2)", link.Code, link.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func resolve(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	var url string
	err := db.QueryRow("SELECT url FROM links WHERE code=$1", code).Scan(&url)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
