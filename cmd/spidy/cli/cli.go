package cli

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/twiny/spidy/v2/cmd/spidy/api"
	"github.com/twiny/spidy/v2/internal/pkg/spider/v1"
	"github.com/urfave/cli/v2"
)

func Exec() {
	app := &cli.App{
		Name:     "Spidy",
		HelpName: "spidy",
		Usage:    "Domain name scraper",
		Version:  api.Version,
		Commands: []*cli.Command{
			initCommand(),
			startCommand(),
			updateCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		return
	}
}

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initialize the spidy configuration",
		Action: func(c *cli.Context) error {
			return initializeConfig()
		},
	}
}
func initializeConfig() error {
	_, err := spider.InitConfig()
	return err
}

func startCommand() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "Start the spidy crawler",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "`path` to config file",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "`path` to output file",
			},
			&cli.BoolFlag{
				Name:    "available",
				Aliases: []string{"a"},
				Usage:   "only output available domains",
			},
		},
		Subcommands: []*cli.Command{
			csvSubcommand(),
			fileSubcommand(),
			stdinSubcommand(),
		},
	}
}

func csvSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "csv",
		Usage: "Read URLs from a CSV file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "`path` to CSV file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "column",
				Aliases:  []string{"c"},
				Usage:    "column `name` to read URLs from",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			file, err := os.Open(c.String("file"))
			if err != nil {
				return fmt.Errorf("failed to open CSV file: %w", err)
			}
			defer file.Close()

			r := csv.NewReader(file)

			header, err := r.Read()
			if err != nil {
				return fmt.Errorf("failed to read header row: %w", err)
			}

			var columnIndex int
			for i, name := range header {
				if name == "column_name" {
					columnIndex = i
					break
				}
			}

			var links []string
			for {
				record, err := r.Read()
				if err != nil {
					break
				}
				links = append(links, record[columnIndex])
			}

			fmt.Println("URLs:", links)

			spidy, err := api.NewSpider(c.String("config"), c.String("output"))
			if err != nil {
				return fmt.Errorf("failed to create spider: %w", err)
			}
			go spidy.Shutdown()

			return spidy.Start(links)
		},
	}
}
func fileSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "file",
		Usage: "Read URLs from a file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "`path` to file",
			},
		},
		Action: func(c *cli.Context) error {
			file, err := os.Open(c.String("file"))
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			var urls []string
			// 2. Read the file line by line.
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				urls = append(urls, line)
				fmt.Println("Received URL:", line)
			}

			// 3. Check for errors.
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading file: %w", err)
			}

			fmt.Println("URLs:", urls)

			time.Sleep(10 * time.Second)

			return nil
		},
	}
}
func stdinSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "stdin",
		Usage: "Read URLs from stdin",
		Action: func(c *cli.Context) error {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Println("Enter URLs (press CTRL+D to finish):")

			var urls []string
			for scanner.Scan() {
				line := scanner.Text()
				urls = append(urls, line)
				fmt.Println("Received URL:", line)
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading from stdin: %v", err)
			}

			fmt.Println("URLs:", urls)

			time.Sleep(10 * time.Second)

			return nil
		},
	}
}

func updateCommand() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update the spidy configuration",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}
