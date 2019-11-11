package crawler

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cockroachlabs/wikifeedia/db"
	"github.com/cockroachlabs/wikifeedia/wikipedia"
	"golang.org/x/sync/errgroup"
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

// CrawlOnce does one pull of the top list of articles and then fetches them all.
func (c *Crawler) CrawlOnce(ctx context.Context) error {
	for _, p := range wikipedia.Projects {
		if err := c.crawlProjectOnce(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (c *Crawler) crawlProjectOnce(ctx context.Context, project string) (err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			fmt.Println("crawl of %s took %v", project, time.Since(start))
		}
	}()
	if err := c.fetchNewTopArticles(ctx, project); err != nil {
		return err
	}
	return c.db.DeleteOldArticles(ctx, project,
		time.Now().UTC().Add(-48*time.Hour).Truncate(24*time.Hour))
}

func (c *Crawler) fetchNewTopArticles(ctx context.Context, project string) error {
	top, err := c.wiki.FetchTopArticles(ctx, project)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	type article struct {
		ta *wikipedia.TopPageviewsArticle
		a  *wikipedia.Article
	}
	const writeConcurrency = 10
	articleChan := make(chan article, writeConcurrency)
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
	writeGroup, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, writeConcurrency)
	for a := range articleChan {
		if a.a.Summary.Extract == "" || len(a.a.Media) == 0 {
			continue
		}
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return writeGroup.Wait()
		}

		writeGroup.Go(func() error {
			defer func() { <-sem }()
			dba := makeArticle(project, a.ta.Views, a.a)
			return c.db.UpsertArticle(ctx, dba)
		})
	}
	return writeGroup.Wait()
}

func makeArticle(project string, pageViews int, a *wikipedia.Article) db.Article {
	dba := db.Article{
		Project:    project,
		Article:    a.Article,
		Title:      a.Summary.Titles.Normalized,
		Abstract:   a.Summary.Extract,
		DailyViews: pageViews,
		ArticleURL: a.Summary.ContentURLs.Desktop.Page,
		Retrieved:  a.Summary.Timestamp,
	}
	if len(a.Media) > 0 {
		dba.ImageURL = a.Media[0].Original.Source
		dba.ThumbnailURL = a.Media[0].Thumbnail.Source
	}
	return dba
}
