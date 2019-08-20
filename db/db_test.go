package db

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var haveCockroach bool

func init() {
	_, err := exec.LookPath("cockroach")
	haveCockroach = err == nil
}

func runCockroach(t *testing.T) (_ *exec.Cmd, pgUrl string, _ error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, "", err
	}
	listenFileName := f.Name()
	f.Close()
	os.Remove(listenFileName)
	defer os.Remove(listenFileName)
	cmd := exec.Command("cockroach", "start", "--insecure",
		"--store=type=mem,size=1G",
		"--listen-addr=localhost:0",
		"--listening-url-file="+listenFileName)
	if err := cmd.Start(); err != nil {
		return nil, "", err
	}
	for {
		if _, err := os.Stat(listenFileName); os.IsNotExist(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	contents, err := ioutil.ReadFile(listenFileName)
	if err != nil {
		return nil, "", err
	}
	pgUrl = strings.TrimSpace(string(contents))
	conf, err := pgx.ParseConnectionString(pgUrl)
	assert.Nil(t, err)
	db, err := pgx.Connect(conf)
	assert.Nil(t, err)
	_, err = db.Exec("CREATE DATABASE " + DatabaseName)
	assert.Nil(t, err)
	assert.Nil(t, db.Close())
	return cmd, pgUrl, nil
}

func TestDB(t *testing.T) {
	if !haveCockroach {
		t.Skip("Don't have cockroach")
	}
	cockroach, pgUrl, err := runCockroach(t)
	require.Nil(t, err)
	defer cockroach.Process.Kill()
	db, err := New(pgUrl)
	require.Nil(t, err)
	articles := []Article{
		{
			Project:    "en",
			Article:    "foo",
			Title:      "foo",
			DailyViews: 123,
		},
		{
			Project:    "en",
			Article:    "bar",
			Title:      "bar",
			DailyViews: 321,
		},
	}
	ctx := context.Background()
	for _, a := range articles {
		assert.Nil(t, db.UpsertArticle(ctx, a))
	}
	got, err := db.GetArticles(ctx, "en", 0, 1000)
	assert.Nil(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, articles[1], got[0])
	assert.Equal(t, articles[0], got[1])
}
