package main

import (
	"context"
	"hash/crc32"
	"os"

	"github.com/pelletier/go-toml/v2"
	"github.com/sintan1729/lure/internal/logger/log"
	"github.com/sintan1729/lure/internal/translate"
	"github.com/urfave/cli/v3"
)

func genIDCmd(ctx context.Context, c *cli.Command) error {
	if c.NArg() < 1 {
		log.Fatal("This command requires one or more arguments").Send()
	}

	fl, err := os.Open(c.Args().First())
	if err != nil {
		log.Fatal("Error opening file").Err(err).Send()
	}

	var tr translate.Translations
	err = toml.NewDecoder(fl).Decode(&tr)
	if err != nil {
		return err
	}
	fl.Close()

	for i, item := range tr.Items {
		tr.Items[i].ID = crc32.ChecksumIEEE([]byte(item.Value))
	}

	fl, err = os.Create(c.Args().First())
	if err != nil {
		log.Fatal("Error creating file").Err(err).Send()
	}

	err = toml.NewEncoder(fl).Encode(tr)
	if err != nil {
		log.Fatal("Error writing toml to file").Err(err).Send()
	}

	return nil
}
