package main

import (
	"flag"
	"fmt"
	"github.com/mackwong/clickhouse-operator/pkg/cmds"
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
					return cmds.InitContainerRun(context)
				},
			},
			{
				Name:        "broker",
				Description: "clickhouse broker for provision、bind...",
				Flags:       cmds.BrokerFlags(),
				Action: func(context *cli.Context) error {
					return cmds.BrokerRun(context)
				},
			},
			{
				Name:        "operator",
				Description: "clickhouse operator",
				Flags:       cmds.OperatorFlags(),
				Action: func(context *cli.Context) error {
					return cmds.OperatorRun(context)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
