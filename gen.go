package main

import (
	"context"
	"os"

	"github.com/sintan1729/lure/pkg/gen"
	"github.com/urfave/cli/v3"
)

var genCmd = &cli.Command{
	Name:    "generate",
	Usage:   "Generate a LURE script from a template",
	Aliases: []string{"gen"},
	Commands: []*cli.Command{
		genPipCmd,
	},
}

var genPipCmd = &cli.Command{
	Name:  "pip",
	Usage: "Generate a LURE script for a pip module",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "version",
			Aliases:  []string{"v"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "description",
			Aliases: []string{"d"},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		return gen.Pip(os.Stdout, gen.PipOptions{
			Name:        c.String("name"),
			Version:     c.String("version"),
			Description: c.String("description"),
		})
	},
}
