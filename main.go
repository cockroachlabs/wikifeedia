// Command wikifeedia does something...
package main

import (
	"fmt"
	"os"

	"github.com/awoods187/wikifeedia/db"
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
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run command: %v\n", err)
		os.Exit(1)
	}
}
