package wikipedia

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const wikimediaURL = "https://wikimedia.org/api/rest_v1"
const wikipediaURL = "https://en.wikipedia.org/api/rest_v1"

type Client struct {
	cli     http.Client
	limiter *rate.Limiter
}

func New() *Client {
	return &Client{
		limiter: rate.NewLimiter(100, 10),
	}
}

type PageViewEntry struct {
	Project string `json:"project"`
	Article string `json:"article`
}

type TopPageviews struct {
	Project  string `json:"project"`
	Access   string `json:"access"`
	Year     string `json:"year"`
	Month    string `json:"month"`
	Day      string `json:"day"`
	Articles []TopPageviewsArticle
}

type TopPageviewsArticle struct {
	Article string `json:"article"`
	Views   int    `json:"views"`
	Rank    int    `json:"rank"`
}

type Article struct {
	Article string
	Summary ArticleSummary
	Media   []ArticleMediaItem
}

type ArticleTitles struct {
	Canonical  string `json:"canonical"`
	Normalized string `json:"normalized"`
	Display    string `json:"display"`
}

type ArticleMediaItem struct {
	SectionID int           `json:"section_id"`
	Type      string        `json:"type"`
	Titles    ArticleTitles `json:"titles"`
	Thumbnail ImageMetadata `json:"thumbnail"`
	Original  ImageMetadata `json:"original"`
}

type ImageMetadata struct {
	Source string `json:"source"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Mime   string `json:"mime"`
}

type ContentURLs struct {
	Desktop ArticleURLs `json:"desktop"`
	Mobile  ArticleURLs `json:"mobile"`
}

type ArticleURLs struct {
	Page string `json:"page"`
}

type ArticleSummary struct {
	Type         string `json:"type"`
	Title        string `json:"title"`
	DisplayTitle string `json:"display_title"`
	Titles       ArticleTitles
	WikibaseItem string      `json:"wikibase_item"`
	Lang         string      `json:"lang"`
	Timestamp    time.Time   `json:"time"`
	Extract      string      `json:"extract"`
	ExtractHTML  string      `json:"extract_html"`
	ContentURLs  ContentURLs `json:"content_urls"`
}

func (c *Client) GetArticle(ctx context.Context, articleName string) (Article, error) {
	fmt.Println(articleName)
	summary, err := c.GetArticleSummary(ctx, articleName)
	if err != nil {
		return Article{}, err
	}
	media, err := c.GetArticleMedia(ctx, articleName)
	if err != nil {
		return Article{}, err
	}
	return Article{
		Article: articleName,
		Summary: summary,
		Media:   media,
	}, nil
}

func (c *Client) GetArticleSummary(
	ctx context.Context, articleName string,
) (summary ArticleSummary, err error) {
	if err = c.limiter.Wait(ctx); err != nil {
		return summary, err
	}
	url := fmt.Sprintf(wikipediaURL + "/page/summary/" + articleName)
	resp, err := c.cli.Get(url)
	if err != nil {
		return summary, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return summary, fmt.Errorf("Unexpected status code %v: resp %s", resp.StatusCode, respBody)
	}
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return ArticleSummary{}, err
	}
	return summary, nil
}

func (c *Client) GetArticleMedia(
	ctx context.Context, articleName string,
) ([]ArticleMediaItem, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	url := fmt.Sprintf(wikipediaURL + "/page/media/" + articleName)
	resp, err := c.cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Items []ArticleMediaItem `json:"items"`
	}
	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Unexpected status code %v: resp %s", resp.StatusCode, respBody)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *Client) FetchTopArticles(ctx context.Context) (*TopPageviews, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	now := time.Now().Add(-48 * time.Hour)
	url := fmt.Sprintf(wikimediaURL+"/metrics/pageviews/top/en.wikipedia.org/all-access/%04d/%02d/%02d",
		now.Year(), int(now.Month()), now.Day())
	resp, err := c.cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Items []TopPageviews `json:"items"`
	}
	if resp.StatusCode != 200 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Unexpected status code %v: resp %s", resp.StatusCode, respBody)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no items found in response")
	}
	results := &result.Items[0]
	results.Articles = filterSpecial(results.Articles)
	return results, nil
}

func shouldFilter(articleName string) bool {
	return strings.HasPrefix(articleName, "Special:") || articleName == "Main_Page"
}

func filterSpecial(top []TopPageviewsArticle) (filtered []TopPageviewsArticle) {
	filtered = top[:0]
	for _, a := range top {
		if !shouldFilter(a.Article) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}
