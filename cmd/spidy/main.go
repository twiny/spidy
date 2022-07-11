package main

import (
	"log"
	"os"

	//

	"github.com/twiny/spidy/v2/cmd/spidy/api"

	//
	"github.com/urfave/cli/v2"
)

// main
func main() {
	app := &cli.App{
		Name:     "Spidy",
		HelpName: "spidy",
		Usage:    "Domain name scraper",
		Version:  api.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "`path` to config file",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "urls",
				Aliases:  []string{"u"},
				Usage:    "`urls` of page to scrape",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			s, err := api.NewSpider(c.String("config"))
			if err != nil {
				return err
			}
			go s.Shutdown()

			return s.Start(c.StringSlice("urls"))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		return
	}
}
