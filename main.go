package main

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/u-speak/configor"
	"github.com/u-speak/core"
	"github.com/u-speak/core/node"
	"github.com/urfave/cli"
)

func main() {
	err := configor.Load(&core.Config, "config.yml", "/etc/uspeak/config.yml")
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
	if core.Config.Logger.Debug {
		log.SetLevel(log.DebugLevel)
	}

	app := cli.NewApp()
	app.Name = "uspeakd"
	app.Version = "0.1.0"
	app.Usage = "Run a uspeak node"
	app.Action = func(c *cli.Context) error {
		if core.Config.Global.SSLKey == "" || core.Config.Global.SSLCert == "" {
			log.Fatal("Could not load SSL Configuration! Since this application handles highly sensitive data, SSL Certificates must be provided")
		}
		log.Infof("Welcome to uspeak!")
		if core.Config.Web.Static.Directory != "false" && core.Config.Web.Static.Directory != "" {
			go core.RunWeb()
		} else {
			log.Info("Static Webserver disabled")
		}

		n, err := node.New(core.Config)
		if err != nil {
			log.Fatal(err)
		}
		go n.Run()
		go core.RunAPI(n)

		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		return nil
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
