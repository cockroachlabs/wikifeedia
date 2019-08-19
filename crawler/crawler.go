package crawler

import (
	"context"
	"fmt"
	"os"

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
	var articles []wikipedia.Article
	failed := map[int]bool{}
	for i := range top.Articles {
		ta := &top.Articles[i]
		article, err := c.wiki.GetArticle(ctx, project, ta.Article)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to retreive %q: %v\n", ta.Article, err)
			failed[i] = true
		}
		articles = append(articles, article)
	}
	for i := range articles {
		if failed[i] {
			continue
		}
		a := &articles[i]
		dba := db.Article{
			Project:    project,
			Article:    a.Article,
			Title:      a.Summary.Titles.Normalized,
			Abstract:   a.Summary.Extract,
			DailyViews: top.Articles[i].Views,
			ArticleURL: a.Summary.ContentURLs.Desktop.Page,
		}
		if len(a.Media) > 0 {
			dba.ImageURL = a.Media[0].Original.Source
			dba.ThumbnailURL = a.Media[0].Thumbnail.Source
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
