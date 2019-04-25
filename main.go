package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

var (
	version = "0.0.0"
	build   = "0"
)

func main() {
	fmt.Println("Inut vars", os.Getenv("PLUGIN_URL"))

	for _, pair := range os.Environ() {
		fmt.Println(pair)
	}
	app := cli.NewApp()
	app.Name = "mattermost message"
	app.Usage = "mattermost message"
	app.Action = run
	app.Version = fmt.Sprintf("%s+%s", version, build)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "url",
			Usage:  "mattermost url",
			EnvVar: "PLUGIN_URL",
		},
		cli.StringFlag{
			Name:   "message",
			Usage:  "Text message to be sent",
			EnvVar: "PLUGIN_MESSAGE",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "Incoming webhook token",
			EnvVar: "PLUGIN_TOKEN",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	fmt.Println("Inside  context", c)

	plugin := Plugin{
		Config: Config{
			Url:     c.String("url"),
			Message: c.String("message"),
			Token:   c.String("token"),
		},
	}

	/*
		if plugin.Config.Url == "" {
			return errors.New("Missing Mattermost URL")
		}

		if plugin.Config.Token == "" {
			return errors.New("Missing Mattermost Token")
		}
	*/

	return plugin.Exec()
}
