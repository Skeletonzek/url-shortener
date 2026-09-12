package main

import (
	"io"
	"math/rand"
	"net/http"
	"sync"
)

const (
	idLength  = 8
	idCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	mu    sync.Mutex
	store = make(map[string]string)
)

func generateID() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = idCharset[rand.Intn(len(idCharset))]
	}
	return string(b)
}

func shortUrlHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "method not allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, "url is required", http.StatusBadRequest)
		return
	}
	originalURL := string(body)

	mu.Lock()
	id := generateID()
	for {
		if _, exists := store[id]; !exists {
			break
		}
		id = generateID()
	}
	store[id] = originalURL
	mu.Unlock()

	shortURL := "http://" + req.Host + "/" + id

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func fullUrlHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "method not allowed", http.StatusBadRequest)
		return
	}

	id := req.PathValue("id")
	if id == "" {
		http.Error(res, "id is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	originalURL, ok := store[id]
	mu.Unlock()

	if !ok {
		http.Error(res, "url not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortUrlHandler)
	mux.HandleFunc(`/{id}`, fullUrlHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
