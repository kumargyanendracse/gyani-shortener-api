package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

var db *sql.DB

func main() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	log.Println("DATABASE_URL length:", len(dbURL))

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("DB ping error:", err)
	}
	http.HandleFunc("/shorten", shorten)
	http.HandleFunc("/resolve", resolve)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	json.NewDecoder(r.Body).Decode(&req)

	code := randString(6)

	_, err := db.Exec("INSERT INTO links (code, url) VALUES ($1,$2)", code, req.URL)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"code": code,
	})
}

func resolve(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	var url string
	err := db.QueryRow("SELECT url FROM links WHERE code=$1", code).Scan(&url)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"url": url,
	})
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
