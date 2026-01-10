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

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

var db *sql.DB

func main() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("DB ping error:", err)
	} else {
		log.Println("Connected to database")
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
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	var req ShortenRequest
	json.NewDecoder(r.Body).Decode(&req)

	if req.URL == "" {
		http.Error(w, "URL is required", 400)
		return
	}

	code := randString(6)

	_, err := db.Exec(
		"INSERT INTO links (code, url) VALUES ($1, $2)",
		code, req.URL,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"code": code,
	})
}

func resolve(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

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
