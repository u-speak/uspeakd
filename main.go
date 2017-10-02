package main

import (
	"os"

	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"github.com/u-speak/core"
	"github.com/urfave/cli"
)

func main() {
	err := configor.Load(&core.Config, "config.yml")
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Name = "uspeakd"
	app.Version = "0.1.0"
	app.Usage = "Run a uspeak node"
	var tempDir string
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "web",
			Usage:       "run a static webserver, serving files in the specified directory. When set to false, the static server is disabled.",
			Destination: &tempDir,
		},
	}
	app.Action = func(c *cli.Context) error {
		if tempDir != "" {
			core.Config.Web.Static.Directory = tempDir
		}
		if core.Config.Web.Static.Directory != "false" && core.Config.Web.Static.Directory != "" {
			go core.RunWeb()
		}
		go core.RunAPI()
		core.RunNode()
		return nil
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
