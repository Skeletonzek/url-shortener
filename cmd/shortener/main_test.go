package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFullURLHandler(t *testing.T) {
	const (
		existingID  = "EwHXdJfB"
		originalURL = "https://practicum.yandex.ru/"
	)
	mu.Lock()
	store[existingID] = originalURL
	mu.Unlock()

	type want struct {
		code     int
		location string
	}
	tests := []struct {
		name   string
		method string
		id     string
		want   want
	}{
		{
			name:   "existing id redirects to original url",
			method: http.MethodGet,
			id:     existingID,
			want:   want{code: http.StatusTemporaryRedirect, location: originalURL},
		},
		{
			name:   "unknown id",
			method: http.MethodGet,
			id:     "unknown1",
			want:   want{code: http.StatusBadRequest},
		},
		{
			name:   "empty id",
			method: http.MethodGet,
			id:     "",
			want:   want{code: http.StatusBadRequest},
		},
		{
			name:   "wrong method POST",
			method: http.MethodPost,
			id:     existingID,
			want:   want{code: http.StatusBadRequest},
		},
		{
			name:   "wrong method DELETE",
			method: http.MethodDelete,
			id:     existingID,
			want:   want{code: http.StatusBadRequest},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/"+tt.id, nil)
			rec := httptest.NewRecorder()

			newRouter().ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.code {
				t.Errorf("status code = %d, want %d", res.StatusCode, tt.want.code)
			}
			if got := res.Header.Get("Location"); got != tt.want.location {
				t.Errorf("Location = %q, want %q", got, tt.want.location)
			}
		})
	}
}

func TestShortURLHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name   string
		method string
		body   string
		want   want
	}{
		{
			name:   "valid url returns short url",
			method: http.MethodPost,
			body:   "https://practicum.yandex.ru/",
			want:   want{code: http.StatusCreated, contentType: "text/plain"},
		},
		{
			name:   "empty body",
			method: http.MethodPost,
			body:   "",
			want:   want{code: http.StatusBadRequest},
		},
		{
			name:   "wrong method GET",
			method: http.MethodGet,
			body:   "https://practicum.yandex.ru/",
			want:   want{code: http.StatusBadRequest},
		},
		{
			name:   "wrong method PUT",
			method: http.MethodPut,
			body:   "https://practicum.yandex.ru/",
			want:   want{code: http.StatusBadRequest},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			newRouter().ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.want.code {
				t.Fatalf("status code = %d, want %d", res.StatusCode, tt.want.code)
			}
			if tt.want.code != http.StatusCreated {
				return
			}

			if got := res.Header.Get("Content-Type"); got != tt.want.contentType {
				t.Errorf("Content-Type = %q, want %q", got, tt.want.contentType)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			prefix := "http://" + req.Host + "/"
			shortURL := string(body)
			if !strings.HasPrefix(shortURL, prefix) {
				t.Fatalf("short url = %q, want prefix %q", shortURL, prefix)
			}
			id := strings.TrimPrefix(shortURL, prefix)
			if len(id) != idLength {
				t.Errorf("id length = %d, want %d", len(id), idLength)
			}

			mu.Lock()
			saved, ok := store[id]
			mu.Unlock()
			if !ok || saved != tt.body {
				t.Errorf("store[%q] = %q (exists: %v), want %q", id, saved, ok, tt.body)
			}
		})
	}
}
