// Package db contains logic for interacting with the database
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx"
)

// Probably I want to create a table for articles and then a separate table for article views

// I want information about articles on an hourly basis as well as a daily basis.

const (
	DatabaseName  = "wikifeedia"
	articlesTable = `CREATE TABLE IF NOT EXISTS articles (
		project STRING NOT NULL,
		article STRING NOT NULL,
		title STRING,
		thumbnail_url STRING,
		image_url STRING,
		abstract STRING,
		article_url STRING,
		daily_views INT NOT NULL,
		retrieved TIMESTAMPTZ NOT NULL,
		INDEX (project, daily_views DESC),
		PRIMARY KEY (project, article)
	);`
)

// Article is the data model for a Wikipedia article.
type Article struct {
	Project      string    `json:"project"`
	Article      string    `json:"article"`
	Title        string    `json:"title"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Abstract     string    `json:"abstract"`
	ImageURL     string    `json:"image_url"`
	ArticleURL   string    `json:"article_url"`
	DailyViews   int       `json:"daily_views"`
	Retrieved    time.Time `json:"retrieved"`
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
			  daily_views,
			  cluster_logical_timestamp()::STRING
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
	getArticlesFollowerReadAsOfSQL = getArticlesSelection +
		` AS OF SYSTEM TIME '%s'` +
		getArticlesModifiers
)

// GetArticles returns the list of articles.
func (db *DB) GetArticles(
	ctx context.Context, project string, offset, limit int, followerRead bool, asOf string,
) (_ []Article, newAsOf string, _ error) {
	stmt := db.getArticles.Name
	if followerRead && asOf == "" {
		stmt = db.getArticlesFollowerRead.Name
	} else if followerRead {
		stmt = fmt.Sprintf(getArticlesFollowerReadAsOfSQL, asOf)
	}
	rows, err := db.connPool.QueryEx(ctx, stmt, nil, project, limit, offset)
	if err != nil {
		return nil, "", err
	}
	var results []Article
	var a Article
	for rows.Next() {
		if err := rows.Scan(&a.Project, &a.Article, &a.Title,
			&a.ThumbnailURL, &a.ImageURL, &a.Abstract,
			&a.ArticleURL, &a.DailyViews, &asOf); err != nil {
			return nil, "", err
		}
		results = append(results, a)
	}
	return results, asOf, rows.Err()
}

// DeleteOldArticles deletes articles which were retrieved before the specified
// time.
func (db *DB) DeleteOldArticles(
	ctx context.Context, project string, retrievedBefore time.Time,
) error {
	_, err := db.connPool.ExecEx(ctx,
		`DELETE FROM articles WHERE project = $1 AND retrieved < $2`,
		nil, project, retrievedBefore)
	return err
}

// UpsertArticle upserts a into the database.
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
				daily_views,
				retrieved
			)
	VALUES
		($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		nil,
		a.Project,
		a.Article,
		a.Title,
		a.ThumbnailURL,
		a.ImageURL,
		a.Abstract,
		a.ArticleURL,
		a.DailyViews,
		a.Retrieved)
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
