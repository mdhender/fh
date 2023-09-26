// Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package config implements a configuration for the Far Horizons game engine and web server.
package config

import (
	"flag"
	"github.com/peterbourgon/ff/v3"
	"os"
)

// Config defines configuration information for the application.
type Config struct {
	Debug      bool
	Home       string
	Host       string
	Port       string
	Public     string
	Templates  string
	WorkingDir string
}

// Default returns a default configuration.
// These are the values without loading the environment, configuration file, or command line.
func Default(home string) (*Config, error) {
	cfg := Config{
		Home:       home,
		Port:       "8080",
		Public:     ".",
		Templates:  ".",
		WorkingDir: ".",
	}

	return &cfg, nil
}

// Load updates the values in a Config in this order:
//  1. It will load a configuration file if one is given on the
//     command line via the `-config` flag. If provided, the file
//     must contain a valid JSON object.
//  2. Environment variables, using the prefix `GOBBS`
//  3. Command line flags
func (cfg *Config) Load() error {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.StringVar(&cfg.Home, "home", cfg.Home, "override HOME path")
	fs.StringVar(&cfg.Host, "host", cfg.Host, "host name (or IP) to bind to")
	fs.StringVar(&cfg.Port, "port", cfg.Port, "port to listen to")
	fs.StringVar(&cfg.Public, "public", cfg.Public, "path to public assets")
	fs.StringVar(&cfg.Templates, "templates", cfg.Templates, "path to template files")
	fs.StringVar(&cfg.WorkingDir, "working-dir", cfg.WorkingDir, "path to run from")

	err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("FH"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser))
	if err != nil {
		return err
	}

	return nil
}
