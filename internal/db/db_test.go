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

package db_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"go.arsenm.dev/lure/internal/db"
)

var testPkg = db.Package{
	Name:    "test",
	Version: "0.0.1",
	Release: 1,
	Epoch:   2,
	Description: db.NewJSON(map[string]string{
		"en": "Test package",
		"ru": "Проверочный пакет",
	}),
	Homepage: db.NewJSON(map[string]string{
		"en": "https://lure.arsenm.dev",
	}),
	Maintainer: db.NewJSON(map[string]string{
		"en": "Arsen Musayelyan <arsen@arsenm.dev>",
		"ru": "Арсен Мусаелян <arsen@arsenm.dev>",
	}),
	Architectures: db.NewJSON([]string{"arm64", "amd64"}),
	Licenses:      db.NewJSON([]string{"GPL-3.0-or-later"}),
	Provides:      db.NewJSON([]string{"test"}),
	Conflicts:     db.NewJSON([]string{"test"}),
	Replaces:      db.NewJSON([]string{"test-old"}),
	Depends: db.NewJSON(map[string][]string{
		"": {"sudo"},
	}),
	BuildDepends: db.NewJSON(map[string][]string{
		"":     {"golang"},
		"arch": {"go"},
	}),
	Repository: "default",
}

func TestInit(t *testing.T) {
	gdb, err := sqlx.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	err = db.Init(gdb, ":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	_, err = gdb.Exec("SELECT * FROM pkgs")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	ver, ok := db.GetVersion(gdb)
	if !ok {
		t.Errorf("Expected version to be present")
	} else if ver != db.CurrentVersion {
		t.Errorf("Expected version %d, got %d", db.CurrentVersion, ver)
	}
}

func TestInsertPackage(t *testing.T) {
	gdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	err = db.InsertPackage(gdb, testPkg)
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	dbPkg := db.Package{}
	err = sqlx.Get(gdb, &dbPkg, "SELECT * FROM pkgs WHERE name = 'test' AND repository = 'default'")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if !reflect.DeepEqual(testPkg, dbPkg) {
		t.Errorf("Expected test package to be the same as database package")
	}
}

func TestGetPkgs(t *testing.T) {
	gdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err = db.InsertPackage(gdb, x1)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = db.InsertPackage(gdb, x2)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	result, err := db.GetPkgs(gdb, "name LIKE 'x%'")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	for result.Next() {
		var dbPkg db.Package
		err = result.StructScan(&dbPkg)
		if err != nil {
			t.Errorf("Expected no error, got %s", err)
		}

		if !strings.HasPrefix(dbPkg.Name, "x") {
			t.Errorf("Expected package name to start with 'x', got %s", dbPkg.Name)
		}
	}
}

func TestGetPkg(t *testing.T) {
	gdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err = db.InsertPackage(gdb, x1)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = db.InsertPackage(gdb, x2)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	pkg, err := db.GetPkg(gdb, "name LIKE 'x%' ORDER BY name")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if pkg.Name != "x1" {
		t.Errorf("Expected x1 package, got %s", pkg.Name)
	}

	if !reflect.DeepEqual(*pkg, x1) {
		t.Errorf("Expected x1 to be %v, got %v", x1, *pkg)
	}
}

func TestDeletePkgs(t *testing.T) {
	gdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"

	err = db.InsertPackage(gdb, x1)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = db.InsertPackage(gdb, x2)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = db.DeletePkgs(gdb, "name = 'x1'")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	var dbPkg db.Package
	err = gdb.Get(&dbPkg, "SELECT * FROM pkgs WHERE name LIKE 'x%' ORDER BY name LIMIT 1;")
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if dbPkg.Name != "x2" {
		t.Errorf("Expected x2 package, got %s", dbPkg.Name)
	}
}

func TestJsonArrayContains(t *testing.T) {
	gdb, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
	defer gdb.Close()

	x1 := testPkg
	x1.Name = "x1"
	x2 := testPkg
	x2.Name = "x2"
	x2.Provides.Val = append(x2.Provides.Val, "x")

	err = db.InsertPackage(gdb, x1)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	err = db.InsertPackage(gdb, x2)
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	var dbPkg db.Package
	err = gdb.Get(&dbPkg, "SELECT * FROM pkgs WHERE json_array_contains(provides, 'x');")
	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	if dbPkg.Name != "x2" {
		t.Errorf("Expected x2 package, got %s", dbPkg.Name)
	}
}
