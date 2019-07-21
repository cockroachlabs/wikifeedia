// Package db contains logic for interacting with the database
package db

import (
	"context"

	"github.com/jackc/pgx"
)

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
		PRIMARY KEY (article)
	);`
)

type Article struct {
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
	return db, nil
}

func (db *DB) GetAllArticles(ctx context.Context) ([]Article, error) {
	rows, err := db.connPool.QueryEx(ctx, `SELECT
		article,
		title,
		thumbnail_url,
		image_url,
		abstract,
		article_url,
		daily_views
	FROM articles ORDER BY daily_views DESC`, nil)
	if err != nil {
		return nil, err
	}
	var results []Article
	var a Article
	for rows.Next() {
		if err := rows.Scan(&a.Article, &a.Title,
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
				article,
				title,
				thumbnail_url,
				image_url,
				abstract,
				article_url,
				daily_views
			)
	VALUES
		($1, $2, $3, $4, $5, $6, $7)`,
		nil,
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
