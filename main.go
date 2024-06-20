package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"stock-news/db/article_service" // Ensure this path is correct and matches your project structure

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	fetcher := &article_service.ArticleFetcherImpl{}
	saver := &article_service.ArticleSaverImpl{}

	articleService := article_service.NewArticleService(fetcher, saver)

	// Use closures to adapt the methods
	r.Post("/api/v1/save-articles", func(w http.ResponseWriter, r *http.Request) {
		var article article_service.Article
		if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := articleService.SaveArticle(article); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})

	r.Get("/api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		sizeStr := r.URL.Query().Get("size")
		size, _ := strconv.Atoi(sizeStr) // Simplified error handling

		articles, err := articleService.FetchArticles(id, size)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(articles)
	})

	http.ListenAndServe(":8080", r)
}
