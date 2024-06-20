package main

import (
	"encoding/json"
	"net/http"
	"stock-news/db/article_service" // Adjust the import path as necessary
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// FetchRequest defines the structure for a fetch request
type FetchRequest struct {
	ID       string
	Size     int
	Response chan FetchResponse // Channel to send the response back
}

// FetchResponse defines the structure for a fetch response
type FetchResponse struct {
	Articles []article_service.Article
	Error    error
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	fetcher := &article_service.ArticleFetcherImpl{}
	saver := &article_service.ArticleSaverImpl{}

	articleService := article_service.NewArticleService(fetcher, saver)

	// Create a channel for save article requests
	saveArticleChan := make(chan article_service.Article)

	// Create a channel for fetch article requests
	fetchArticleChan := make(chan FetchRequest)

	// Start a goroutine to process save requests
	go func() {
		for article := range saveArticleChan {
			if err := articleService.SaveArticle(article); err != nil {
				// Handle error - Note: This is running in a separate goroutine
				// You might want to log this error or handle it appropriately
			}
			// Placeholder statement to indicate the branch is intentionally empty
			// TODO: Handle error or log it appropriately
			_ = article
		}
	}()

	// Start a goroutine to process fetch requests
	go func() {
		for req := range fetchArticleChan {
			articles, err := articleService.FetchArticles(req.ID, req.Size)
			req.Response <- FetchResponse{Articles: articles, Error: err}
		}
	}()

	r.Post("/api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		var article article_service.Article
		if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Send the article to the save channel instead of saving directly
		saveArticleChan <- article

		w.WriteHeader(http.StatusCreated)
	})

	r.Get("/api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		sizeStr := r.URL.Query().Get("size")
		size, _ := strconv.Atoi(sizeStr) // Simplified error handling

		responseChan := make(chan FetchResponse)
		fetchArticleChan <- FetchRequest{ID: id, Size: size, Response: responseChan}

		// Wait for the response
		response := <-responseChan
		if response.Error != nil {
			http.Error(w, response.Error.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.Articles)
	})

	http.ListenAndServe(":8080", r)
}
