package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
)

func TestGetChiArticles(t *testing.T) {
	// Create a new request with a dummy ID and size
	req, err := http.NewRequest("GET", "/api/v1/articles?id=dummy&size=5", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to record the response
	rr := httptest.NewRecorder()

	// Create a new router and handle the request
	r := chi.NewRouter()
	r.Get("/api/v1/articles", getChiArticles)
	r.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d but got %d", http.StatusOK, rr.Code)
	}

	// Check the response content type
	expectedContentType := "application/json"
	if rr.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("expected content type %s but got %s", expectedContentType, rr.Header().Get("Content-Type"))
	}

	// Decode the response body into an articleResponse struct
	var response articleResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	// Check the number of articles in the response
	expectedSize := 5
	if len(response.Data) != expectedSize {
		t.Errorf("expected %d articles but got %d", expectedSize, len(response.Data))
	}

	// Add additional assertions for the response data if needed
}
