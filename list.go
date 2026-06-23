/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Elara Musayelyan
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"fmt"

	"github.com/sintan1729/lure/internal/config"
	"github.com/sintan1729/lure/internal/db"
	"github.com/sintan1729/lure/pkg/loggerctx"
	"github.com/sintan1729/lure/pkg/manager"
	"github.com/sintan1729/lure/pkg/repos"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slices"
)

var listCmd = &cli.Command{
	Name:    "list",
	Usage:   "List LURE repo packages",
	Aliases: []string{"ls"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "installed",
			Aliases: []string{"I"},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		log := loggerctx.From(ctx)

		err := repos.Pull(ctx, config.Config(ctx).Repos)
		if err != nil {
			log.Fatal("Error pulling repositories").Err(err).Send()
		}

		where := "true"
		args := []any(nil)
		if c.NArg() > 0 {
			where = "name LIKE ? OR json_array_contains(provides, ?)"
			args = []any{c.Args().First(), c.Args().First()}
		}

		result, err := db.GetPkgs(ctx, where, args...)
		if err != nil {
			log.Fatal("Error getting packages").Err(err).Send()
		}
		defer result.Close()

		var installed map[string]string
		if c.Bool("installed") {
			mgr := manager.Detect()
			if mgr == nil {
				log.Fatal("Unable to detect a supported package manager on the system").Send()
			}

			installed, err = mgr.ListInstalled(&manager.Opts{AsRoot: false})
			if err != nil {
				log.Fatal("Error listing installed packages").Err(err).Send()
			}
		}

		for result.Next() {
			var pkg db.Package
			err := result.StructScan(&pkg)
			if err != nil {
				return err
			}

			if slices.Contains(config.Config(ctx).IgnorePkgUpdates, pkg.Name) {
				continue
			}

			version := pkg.Version
			if c.Bool("installed") {
				instVersion, ok := installed[pkg.Name]
				if !ok {
					continue
				} else {
					version = instVersion
				}
			}

			fmt.Printf("%s/%s %s\n", pkg.Repository, pkg.Name, version)
		}

		if err != nil {
			log.Fatal("Error iterating over packages").Err(err).Send()
		}

		return nil
	},
}
