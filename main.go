package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
)

type Article struct {
	Symbol    string `json:"symbol"`
	CreatedAt string `json:"created_at"`
	Headline  string `json:"headline"`
}

type ArticleFetcher interface {
	FetchArticles(id string, size int) ([]Article, error)
}

type ArticleFetcherImpl struct{}

type articleResponse struct {
	Data []Article `json:"data"`
}

func (a *ArticleFetcherImpl) FetchArticles(id string, size int) ([]Article, error) {
	return FetchArticles(id, size)
}

type ArticlwSave interface {
	SaveArticle(article Article) error
}

type ArticleSaveImpl struct{}

func (a *ArticleSaveImpl) SaveArticle(article Article) error {
	return SaveArticle(article)
}

func SaveArticle(article Article) error {
	// Fetch articles from database
	file, err := os.ReadFile("db/articles.json")
	if err != nil {
		return err
	}

	var allArticles []Article

	// Decode the JSON file into the articles slice
	if err := json.Unmarshal(file, &allArticles); err != nil {
		return err
	}

	allArticles = append(allArticles, article)

	// Encode the articles slice into a JSON file
	data, err := json.Marshal(allArticles)
	if err != nil {
		return err
	}

	if err := os.WriteFile("db/articles.json", data, 0644); err != nil {
		return err
	}

	return nil
}

func FetchArticles(id string, size int) ([]Article, error) {
	// Fetch articles from database
	file, err := os.ReadFile("db/articles.json")
	if err != nil {
		return nil, err
	}

	var allArticles []Article

	// Decode the JSON file into the articles slice
	if err := json.Unmarshal(file, &allArticles); err != nil {
		return nil, err
	}

	// Filter the articles based on the id
	var filteredArticles []Article

	for _, article := range allArticles {
		if article.Symbol == id {
			filteredArticles = append(filteredArticles, article)
		}
	}

	artCount := len(filteredArticles)

	if artCount > 0 {
		if artCount < size {
			size = artCount
		}
		filteredArticles = filteredArticles[:size]
	}

	return filteredArticles, nil
}

func main() {
	// Create a new router
	r := chi.NewRouter()

	// Create a new article fetcher
	articleFetcher := &ArticleFetcherImpl{}
	articleSaver := &ArticleSaveImpl{}
	// Add a new route to the router
	r.Get("/api/v1/articles", getChiArticles(articleFetcher))
	r.Post("/api/v1/save-articles", saveChiArticle(articleSaver))

	fmt.Println("Server is running on port 8080")
	// Start the HTTP server
	http.ListenAndServe(":8080", r)
}

func saveChiArticle(articleSaver ArticlwSave) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body
		var article Article
		if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Save the article
		if err := articleSaver.SaveArticle(article); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write a success response
		w.Write([]byte("Article saved successfully"))
	}
}

func getChiArticles(articleFetcher ArticleFetcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the query parameters
		id := r.URL.Query().Get("id")
		size := 5
		if s := r.URL.Query().Get("size"); s != "" {
			size = 5
		}

		// Fetch the articles
		articles, err := articleFetcher.FetchArticles(id, size)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a response
		response := articleResponse{Data: articles}

		// Write the response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
