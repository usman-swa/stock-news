package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"stock-news/db/article_service"
	"strconv"
	"sync"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type FetchRequest struct {
	ID       string
	Size     int
	Response chan FetchResponse
}

type FetchResponse struct {
	Articles []article_service.Article
	Error    error
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)

	fetcher := &article_service.ArticleFetcherImpl{}
	saver := &article_service.ArticleSaverImpl{}
	articleService := article_service.NewArticleService(fetcher, saver)

	saveArticleChan := make(chan article_service.Article)
	fetchArticleChan := make(chan FetchRequest)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for article := range saveArticleChan {
			if err := articleService.SaveArticle(article); err != nil {
				fmt.Println("Error saving article:", err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for req := range fetchArticleChan {
			articles, err := articleService.FetchArticles(req.ID, req.Size)
			req.Response <- FetchResponse{Articles: articles, Error: err}
		}
	}()

	setupHTTPHandlers(r, saveArticleChan, fetchArticleChan)

	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("HTTP server Shutdown: %v", err)
	}

	close(saveArticleChan)
	close(fetchArticleChan)
	wg.Wait()
}

func setupHTTPHandlers(r *chi.Mux, saveArticleChan chan article_service.Article, fetchArticleChan chan FetchRequest) {
	r.Post("/api/v1/save-articles", func(w http.ResponseWriter, r *http.Request) {
		var article article_service.Article
		if err := json.NewDecoder(r.Body).Decode(&article); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		saveArticleChan <- article
		w.WriteHeader(http.StatusCreated)
	})

	r.Get("/api/v1/articles", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		sizeStr := r.URL.Query().Get("size")
		size, _ := strconv.Atoi(sizeStr)

		responseChan := make(chan FetchResponse)
		fetchArticleChan <- FetchRequest{ID: id, Size: size, Response: responseChan}

		response := <-responseChan
		if response.Error != nil {
			http.Error(w, response.Error.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response.Articles)
	})
}
