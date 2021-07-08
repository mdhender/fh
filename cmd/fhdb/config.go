/*
 * FH - Far Horizons server
 * Copyright (c) 2021  Michael D Henderson
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"github.com/peterbourgon/ff/v3"
	"os"
	"path/filepath"
)

type Config struct {
	Debug bool
	Data  struct {
		JDB    string // path to json data file
		Orders string // path to orders files
	}
}

// DefaultConfig returns a default configuration.
// These are the values without loading the environment, configuration file, or command line.
func DefaultConfig() *Config {
	var cfg Config
	cfg.Data.JDB = "D:\\GoLand\\fh\\testdata\\gamma"
	cfg.Data.Orders = "D:\\GoLand\\fh\\testdata\\gamma"
	return &cfg
}

// Load updates the values in a Config in this order:
//   1. It will load a configuration file if one is given on the
//      command line via the `-config` flag. If provided, the file
//      must contain a valid JSON object.
//   2. Environment variables, using the prefix `CONDUIT_RYER_SERVER`
//   3. Command line flags
func (cfg *Config) Load() error {
	fs := flag.NewFlagSet("Server", flag.ExitOnError)
	debug := fs.Bool("debug", cfg.Debug, "log debug information (optional)")
	dataJDB := fs.String("jdb-path", cfg.Data.JDB, "path to json data")
	dataOrders := fs.String("data", cfg.Data.Orders, "path to orders files")

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("FHDB"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser)); err != nil {
		return err
	}

	cfg.Debug = *debug
	cfg.Data.JDB = filepath.Clean(*dataJDB)
	cfg.Data.Orders = filepath.Clean(*dataOrders)

	return nil
}
