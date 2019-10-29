// Package db contains logic for interacting with the database
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx"
)

// Probably I want to create a table for articles and then a separate table for article views

// I want information about articles on an hourly basis as well as a daily basis.

const (
	DatabaseName  = "wikifeedia"
	articlesTable = `CREATE TABLE IF NOT EXISTS articles (
		project STRING,
		article STRING,
		title STRING,
		thumbnail_url STRING,
		image_url STRING,
		abstract STRING,
		article_url STRING,
		daily_views INT,
		INDEX (project, daily_views DESC),
		PRIMARY KEY (project, article)
	);`
)

type Article struct {
	Project      string `json:"project"`
	Article      string `json:"article"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Abstract     string `json:"abstract"`
	ImageURL     string `json:"image_url"`
	ArticleURL   string `json:"article_url"`
	DailyViews   int    `json:"daily_views"`
}

// DB is a wrapper around a pgx.ConnPool that knows about the structure of our application schema.
type DB struct {
	connPool *pgx.ConnPool
	conf     pgx.ConnPoolConfig

	getArticles             *pgx.PreparedStatement
	getArticlesFollowerRead *pgx.PreparedStatement
}

// MaxConnections controls the maximum number of connections for a DB.
const MaxConnections = 256

// New creates a new DB.
func New(pgurl string) (*DB, error) {
	conf, err := pgx.ParseConnectionString(pgurl)
	if err != nil {
		return nil, err
	}
	conf.Database = DatabaseName
	poolConf := pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: MaxConnections,
	}
	connPool, err := pgx.NewConnPool(poolConf)
	if err != nil {
		return nil, err
	}
	db := &DB{
		conf:     poolConf,
		connPool: connPool,
	}
	if err := setupDatabase(db); err != nil {
		return nil, err
	}
	db.getArticles, err = connPool.Prepare("get_articles", getArticlesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare get_articles: %v", err)
	}
	db.getArticlesFollowerRead, err = connPool.Prepare("get_articles_follower_read", getArticlesFollowerReadSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare get_articles_follower_read: %v", err)
	}
	return db, nil
}

const (
	getArticlesSelection = `SELECT * FROM (
  SELECT
    project,
		article,
		title,
		thumbnail_url,
		image_url,
		abstract,
		article_url,
		daily_views
	FROM articles
  WHERE project = $1
  ORDER BY daily_views DESC
  LIMIT ($2 + $3)
)`
	getArticlesModifiers       = `ORDER BY daily_views DESC OFFSET $3`
	getArticlesSQL             = getArticlesSelection + getArticlesModifiers
	getArticlesFollowerReadSQL = getArticlesSelection +
		` AS OF SYSTEM TIME experimental_follower_read_timestamp()` +
		getArticlesModifiers
)

func (db *DB) GetArticles(
	ctx context.Context, project string, offset, limit int, followerRead bool,
) ([]Article, error) {
	stmt := db.getArticles
	if followerRead {
		stmt = db.getArticlesFollowerRead
	}
	rows, err := db.connPool.QueryEx(ctx, stmt.Name, nil, project, limit, offset)
	if err != nil {
		return nil, err
	}
	var results []Article
	var a Article
	for rows.Next() {
		if err := rows.Scan(&a.Project, &a.Article, &a.Title,
			&a.ThumbnailURL, &a.ImageURL, &a.Abstract,
			&a.ArticleURL, &a.DailyViews); err != nil {
			return nil, err
		}
		results = append(results, a)
	}
	return results, nil
}

func (db *DB) UpsertArticle(ctx context.Context, a Article) error {
	_, err := db.connPool.ExecEx(ctx, `UPSERT
	INTO
		articles
			(
        project,
				article,
				title,
				thumbnail_url,
				image_url,
				abstract,
				article_url,
				daily_views
			)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)`,
		nil,
		a.Project,
		a.Article,
		a.Title,
		a.ThumbnailURL,
		a.ImageURL,
		a.Abstract,
		a.ArticleURL,
		a.DailyViews)
	return err
}

func setupDatabase(db *DB) error {
	if _, err := db.connPool.Exec("CREATE DATABASE IF NOT EXISTS " + DatabaseName); err != nil {
		return err
	}
	_, err := db.connPool.Exec(articlesTable)
	// TODO(ajwerner): validate that the table that exists matches what we expect.
	return err
}
