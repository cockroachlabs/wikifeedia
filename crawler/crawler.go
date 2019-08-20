package crawler

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/cockroachlabs/wikifeedia/db"
	"github.com/cockroachlabs/wikifeedia/wikipedia"
)

type Crawler struct {
	db   *db.DB
	wiki *wikipedia.Client
}

// New creates a new crawler.
func New(db *db.DB, wiki *wikipedia.Client) *Crawler {
	return &Crawler{
		db:   db,
		wiki: wiki,
	}
}

func (c *Crawler) crawlProjectOnce(ctx context.Context, project string) error {
	top, err := c.wiki.FetchTopArticles(ctx, project)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	type article struct {
		ta *wikipedia.TopPageviewsArticle
		a  *wikipedia.Article
	}
	articleChan := make(chan article, 10)
	fetchArticle := func(ta *wikipedia.TopPageviewsArticle) {
		defer wg.Done()
		a, err := c.wiki.GetArticle(ctx, project, ta.Article)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to retreive %q: %v\n", ta.Article, err)
			return
		}
		articleChan <- article{
			a:  &a,
			ta: ta,
		}
	}
	for i := range top.Articles {
		wg.Add(1)
		go fetchArticle(&top.Articles[i])
	}
	go func() { wg.Wait(); close(articleChan) }()
	for a := range articleChan {
		if a.a.Summary.Extract == "" || len(a.a.Media) == 0 {
			continue
		}
		dba := db.Article{
			Project:    project,
			Article:    a.a.Article,
			Title:      a.a.Summary.Titles.Normalized,
			Abstract:   a.a.Summary.Extract,
			DailyViews: a.ta.Views,
			ArticleURL: a.a.Summary.ContentURLs.Desktop.Page,
		}
		if len(a.a.Media) > 0 {
			dba.ImageURL = a.a.Media[0].Original.Source
			dba.ThumbnailURL = a.a.Media[0].Thumbnail.Source
		}
		if err := c.db.UpsertArticle(ctx, dba); err != nil {
			return err
		}
	}
	return nil
}

// CrawlOnce does one pull of the top list of articles and then fetches them all.
func (c *Crawler) CrawlOnce(ctx context.Context) error {
	for _, p := range wikipedia.Projects {
		if err := c.crawlProjectOnce(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
