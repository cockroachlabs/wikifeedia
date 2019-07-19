// Package db contains logic for interacting with the database
package db

import (
	"github.com/jackc/pgx"
)

const (
	databaseName  = "wikifeedia"
	articlesTable = `CREATE TABLE IF NOT EXISTS wikifeedia.articles (
		project STRING,
		article STRING,
		image_url STRING,
		summary STRING,
		monthly_views INT,
		daily_views INT,
		hotness FLOAT8,
		PRIMARY KEY (project, article)
	);`
)

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

func setupDatabase(db *DB) error {
	if _, err := db.connPool.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName); err != nil {
		return err
	}
	_, err := db.connPool.Exec(articlesTable)
	// TODO(ajwerner): validate that the table that exists matches what we expect.
	return err
}
