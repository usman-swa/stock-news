package article_service

import (
	"encoding/json"
	"os"
)

type Article struct {
	Symbol    string `json:"symbol"`
	CreatedAt string `json:"created_at"`
	Headline  string `json:"headline"`
}

type ArticleService struct {
	articleFetcher ArticleFetcher
	articleSaver   ArticleSaver
}

// ArticleFetcher interface
type ArticleFetcher interface {
	FetchArticles(id string, size int) ([]Article, error)
}

type ArticleFetcherImpl struct{}

func (a *ArticleFetcherImpl) FetchArticles(id string, size int) ([]Article, error) {
	return FetchArticles(id, size)
}

// ArticleSaver interface
type ArticleSaver interface {
	SaveArticle(article Article) error
}

type ArticleSaverImpl struct{}

func (a *ArticleSaverImpl) SaveArticle(article Article) error {
	return SaveArticle(article)
}

func NewArticleService(fetcher ArticleFetcher, saver ArticleSaver) *ArticleService {
	return &ArticleService{
		articleFetcher: fetcher,
		articleSaver:   saver,
	}
}

func (a *ArticleService) FetchArticles(id string, size int) ([]Article, error) {
	return a.articleFetcher.FetchArticles(id, size)
}

func (a *ArticleService) SaveArticle(article Article) error {
	return a.articleSaver.SaveArticle(article)
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
