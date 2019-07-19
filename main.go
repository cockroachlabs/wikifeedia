// Command wikifeedia does something...
package main

import (
	"os"

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
	app.Run(os.Args)
}
