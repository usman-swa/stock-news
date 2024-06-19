package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

// attributes represents the attributes of an article.
type attributes struct {
	PublishOn time.Time `json:"publishOn"`
	Title     string    `json:"title"`
}

// dataItem represents a single item in the article response data.
type dataItem struct {
	Attributes attributes `json:"attributes"`
}

// articleResponse represents the response structure for the articles API.
type articleResponse struct {
	Data []dataItem `json:"data"`
}

// main is the entry point of the application.
func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/api/v1/articles", getChiArticles)
	r.Post("/news/v2/save-article", saveChiArticle)

	r.Post("/post", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	fmt.Println("Server is running on port 3000")
	http.ListenAndServe(":3000", r)
}

// getChiArticles is the handler function for the "/api/v1/articles" endpoint.
// It retrieves articles based on the provided ID and size parameters.
func getChiArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Fetching Articles..."))

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	sizeStr := r.URL.Query().Get("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		http.Error(w, "Invalid size value", http.StatusBadRequest)
		return
	}
	if size == 0 {
		size = 10 // Default size
	}

	var wg sync.WaitGroup
	var articles []Article
	errChan := make(chan error, 1) // Buffered channel for error handling

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		articles, err = getArticles(id, size)
		if err != nil {
			errChan <- err
			return
		}
	}()

	wg.Wait()      // Wait for the Go routine to finish
	close(errChan) // Close the channel

	if err, ok := <-errChan; ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Process articles and respond
	// Assuming processing and response code goes here

	var articleResponse articleResponse

	for _, article := range articles {
		articleResponse.Data = append(articleResponse.Data, dataItem{
			Attributes: attributes{
				PublishOn: article.CreatedAt,
				Title:     article.Headline,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(articleResponse)
}

// saveChiArticle is the handler function for the "/news/v2/save-article" endpoint.
// It saves the provided article to the database.
func saveChiArticle(w http.ResponseWriter, r *http.Request) {
	var article Article // Assuming Article is a struct representing your article data

	// Decode JSON from request body
	err := json.NewDecoder(r.Body).Decode(&article)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = saveArticle(article)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(article)
}

// Article represents an article with symbol, creation date, and headline.
type Article struct {
	Symbol    string
	CreatedAt time.Time
	Headline  string
}

// getArticles retrieves articles from the database based on the provided ID and size.
func getArticles(id string, size int) ([]Article, error) {
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

// saveArticle saves the provided article to the database.
func saveArticle(article Article) error {
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
	articlesJSON, err := json.Marshal(allArticles)
	if err != nil {
		return err
	}

	// Write the JSON file to the disk
	if err := os.WriteFile("db/articles.json", articlesJSON, 0644); err != nil {
		return err
	}

	return nil
}
