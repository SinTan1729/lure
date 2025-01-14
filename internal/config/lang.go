/*
 * LURE - Linux User REpository
 * Copyright (C) 2023 Arsen Musayelyan
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

package config

import (
	"os"
	"strings"

	"go.arsenm.dev/logger/log"
	"golang.org/x/text/language"
)

var Language language.Tag

func init() {
	lang := SystemLang()
	tag, err := language.Parse(lang)
	if err != nil {
		log.Fatal("Error parsing system language").Err(err).Send()
	}
	base, _ := tag.Base()
	Language = language.Make(base.String())
}

func SystemLang() string {
	lang := os.Getenv("LANG")
	lang, _, _ = strings.Cut(lang, ".")
	if lang == "" {
		lang = "en"
	}
	return lang
}
