// Package article_service provides functionality for fetching and saving articles.
package article_service

import (
	"encoding/json"
	"os"
)

// Article represents a news article.
type Article struct {
	Symbol    string `json:"symbol"`
	CreatedAt string `json:"created_at"`
	Headline  string `json:"headline"`
}

// ArticleService is a service that interacts with the article data.
type ArticleService struct {
	articleFetcher ArticleFetcher
	articleSaver   ArticleSaver
}

// ArticleFetcher is an interface for fetching articles.
type ArticleFetcher interface {
	FetchArticles(id string, size int) ([]Article, error)
}

// ArticleFetcherImpl is an implementation of the ArticleFetcher interface.
type ArticleFetcherImpl struct{}

// FetchArticles fetches articles based on the given ID and size.
func (a *ArticleFetcherImpl) FetchArticles(id string, size int) ([]Article, error) {
	return FetchArticles(id, size)
}

// ArticleSaver is an interface for saving articles.
type ArticleSaver interface {
	SaveArticle(article Article) error
}

// ArticleSaverImpl is an implementation of the ArticleSaver interface.
type ArticleSaverImpl struct{}

// SaveArticle saves the given article.
func (a *ArticleSaverImpl) SaveArticle(article Article) error {
	return SaveArticle(article)
}

// NewArticleService creates a new instance of ArticleService with the provided fetcher and saver.
func NewArticleService(fetcher ArticleFetcher, saver ArticleSaver) *ArticleService {
	return &ArticleService{
		articleFetcher: fetcher,
		articleSaver:   saver,
	}
}

// FetchArticles fetches articles based on the given ID and size using the article fetcher.
func (a *ArticleService) FetchArticles(id string, size int) ([]Article, error) {
	return a.articleFetcher.FetchArticles(id, size)
}

// SaveArticle saves the given article using the article saver.
func (a *ArticleService) SaveArticle(article Article) error {
	return a.articleSaver.SaveArticle(article)
}

// SaveArticle saves the given article to the database.
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

// FetchArticles fetches articles from the database based on the given ID and size.
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

	var filteredArticles []Article

	for _, article := range allArticles {
		if article.Symbol == id {
			filteredArticles = append(filteredArticles, article)
		}
	}

	var artCount = len(filteredArticles)

	if artCount > 0 {
		if artCount < size {
			size = artCount
		}

		filteredArticles = filteredArticles[:size]
	}

	return filteredArticles, nil
}
