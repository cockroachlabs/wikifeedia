// Command wikifeedia does something...
package main

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/cockroachlabs/wikifeedia/crawler"
	"github.com/cockroachlabs/wikifeedia/db"
	"github.com/cockroachlabs/wikifeedia/server"
	"github.com/cockroachlabs/wikifeedia/wikipedia"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "wikifeedia"
	app.Usage = "runs one of the main actions"
	app.Action = func(c *cli.Context) error {
		println("run a subcommand")
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "pgurl",
			Value: "pgurl://root@localhost:26257?sslmode=disable",
		},
	}
	app.Commands = []cli.Command{
		{
			Name: "setup",
			Action: func(c *cli.Context) error {
				pgurl := c.GlobalString("pgurl")
				fmt.Println("Setting up database at", pgurl)
				_, err := db.New(pgurl)
				return err
			},
		},
		{
			Name:        "crawl",
			Description: "Update the set of articles one time",
			Action: func(c *cli.Context) error {
				pgurl := c.GlobalString("pgurl")
				fmt.Println("Setting up database at", pgurl)
				conn, err := db.New(pgurl)
				if err != nil {
					return err
				}
				wiki := wikipedia.New()
				crawl := crawler.New(conn, wiki)
				return crawl.CrawlOnce(context.Background())
			},
		},
		{
			Name:        "server",
			Description: "Run the server",
			Action: func(c *cli.Context) error {
				pgurl := c.GlobalString("pgurl")
				fmt.Println("Setting up database at", pgurl)
				conn, err := db.New(pgurl)
				if err != nil {
					return err
				}
				h := server.New(conn)
				server := http.Server{
					Addr:    fmt.Sprintf(":%d", c.Int("port")),
					Handler: h,
				}
				if !c.Bool("insecure") {
					priv, certBytes, err := generateCertificate()
					if err != nil {
						return errors.Wrapf(err, "failed to generate certificate")
					}
					server.TLSConfig = &tls.Config{
						Certificates: []tls.Certificate{{
							Certificate: [][]byte{certBytes},
							PrivateKey:  priv,
						}},
					}
					if err := server.ListenAndServeTLS("" /* certfile */, "" /* keyfile */); err != nil && err != http.ErrServerClosed {
						return errors.Wrap(err, "failed to start server")
					}
				} else {
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						return errors.Wrap(err, "failed to start server")
					}
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "port",
					Value: 8080,
					Usage: "port on which to serve",
				},
				cli.BoolFlag{
					Name:  "insecure",
					Usage: "disables TLS",
				},
			},
		},
		{
			Name:        "fetch-top-articles",
			Description: "debug command to exercise the wikipedia client functionality.",
			Action: func(c *cli.Context) error {
				ctx := context.Background()
				wiki := wikipedia.New()
				project := c.String("project")
				top, err := wiki.FetchTopArticles(ctx, project)
				if err != nil {
					return err
				}
				n := c.Int("num-articles")
				for i := 0; i < len(top.Articles) && i < n; i++ {
					article, err := wiki.GetArticle(ctx, project, top.Articles[i].Article)
					if err != nil {
						return err
					}
					if i > 0 {
						fmt.Println()
					}
					fmt.Printf("%d. %s (%d)\n\n%s\n", i+1, article.Summary.Titles.Normalized, top.Articles[i].Views, article.Summary.Extract)
				}
				return nil
			},
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "num-articles,n",
					Value: 10,
					Usage: "number of articles to fetch",
				},
				cli.StringFlag{
					Name:  "project",
					Value: "en",
					Usage: "project to scan",
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %v\n", err)
		os.Exit(1)
	}
}

func generateCertificate() (crypto.PrivateKey, []byte, error) {
	// Loosely based on https://golang.org/src/crypto/tls/generate_cert.go
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate private key")
	}

	now := time.Now().UTC()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate serial number")
	}

	cert := x509.Certificate{
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0),
		SerialNumber:          serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cockroach Labs"},
		},
	}

	bytes, err := x509.CreateCertificate(rand.Reader, &cert, &cert, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate certificate")
	}
	return priv, bytes, err
}
