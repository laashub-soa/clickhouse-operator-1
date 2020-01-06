package main

import (
	"flag"
	"fmt"
	broker "github.com/mackwong/clickhouse-operator/cmd/broker"
	init_container "github.com/mackwong/clickhouse-operator/cmd/init-container"
	"github.com/mackwong/clickhouse-operator/cmd/manager"
	"github.com/mackwong/clickhouse-operator/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	flag.Parse()

	cli.VersionFlag = &cli.BoolFlag{
		Name: "print-version", Aliases: []string{"V"},
		Usage: "print only the version",
	}

	app := &cli.App{
		Name:    "clickhouse tools",
		Version: version.Version,
		Authors: []*cli.Author{
			{
				Name:  "Wang Jun",
				Email: "wangjun3@sensetime.com",
			},
		},
		Description: "",
		Usage:       "make an explosive entrance",
		Action: func(c *cli.Context) error {
			fmt.Println("please use help for more information")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "init",
				Description: "prepare clickhouse server environment",
				Action: func(context *cli.Context) error {
					return init_container.Run(context)
				},
			},
			{
				Name:        "broker",
				Description: "clickhouse broker for provision、bind...",
				Flags:       broker.Flags(),
				Action: func(context *cli.Context) error {
					return broker.Run(context)
				},
			},
			{
				Name:        "operator",
				Description: "clickhouse operator",
				Action: func(context *cli.Context) error {
					return manager.Run(context)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
