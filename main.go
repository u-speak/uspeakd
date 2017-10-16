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

	switch core.Config.Logger.Format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
		log.Info("Using json formatter")
	default:
		log.Info("Using default formatter")
	}
	app := cli.NewApp()
	app.Name = "uspeakd"
	app.Version = "0.1.0"
	app.Usage = "Run a uspeak node"
	app.Action = func(c *cli.Context) error {
		log.Infof("Welcome to uspeak!")
		if core.Config.Web.Static.Directory != "false" && core.Config.Web.Static.Directory != "" {
			go core.RunWeb()
		} else {
			log.Info("Static Webserver disabled")
		}
		core.Run()
		return nil
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
