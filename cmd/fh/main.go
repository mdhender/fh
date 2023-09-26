// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"github.com/mdhender/fh/internal/config"
	"github.com/mdhender/fh/internal/dot"
	"github.com/mdhender/fh/internal/homedir"
	"github.com/mdhender/fh/internal/server"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	defer func(started time.Time) {
		log.Printf("[main] elapsed time %v\n", time.Now().Sub(started))
	}(time.Now())
	log.Println("[main] entered")

	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("home: %v\n", err)
	}

	if err := dot.Load("FH", false, false); err != nil {
		log.Fatalf("main: %+v\n", err)
	}

	cfg, err := config.Default(home)
	if err != nil {
		log.Fatal(err)
	} else if err = cfg.Load(); err != nil {
		log.Fatal(err)
	} else if cfg.WorkingDir, err = filepath.Abs(cfg.WorkingDir); err != nil {
		log.Fatalf("[fh] working-dir: %v\n", err)
	} else if sb, err := os.Stat(cfg.WorkingDir); err != nil {
		log.Fatalf("[fh] working-dir: %v\n", err)
	} else if !sb.IsDir() {
		log.Fatalf("[fh] working-dir: invalid path %q\n", cfg.WorkingDir)
	} else if err = os.Chdir(cfg.WorkingDir); err != nil {
		log.Fatalf("[fh] working-dir: %v\n", err)
	} else if wd, err := os.Getwd(); err != nil {
		log.Fatalf("[fh] working-dir: %v\n", err)
	} else {
		log.Printf("[main] working dir %q\n", wd)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config.Config) error {
	var options []server.Option
	var err error

	if cfg.Public, err = filepath.Abs(cfg.Public); err != nil {
		log.Fatalf("[fh] public: %v\n", err)
	} else if sb, err := os.Stat(cfg.Public); err != nil {
		log.Fatalf("[fh] public: %v\n", err)
	} else if !sb.IsDir() {
		log.Fatalf("[fh] public: invalid path %q\n", cfg.Public)
	}
	options = append(options, server.WithAssets("public", cfg.Public))

	if cfg.Templates, err = filepath.Abs(cfg.Templates); err != nil {
		log.Fatalf("[fh] templates: %v\n", err)
	} else if sb, err := os.Stat(cfg.Templates); err != nil {
		log.Fatalf("[fh] templates: %v\n", err)
	} else if !sb.IsDir() {
		log.Fatalf("[fh] templates: invalid path %q\n", cfg.Templates)
	}
	options = append(options, server.WithAssets("templates", cfg.Templates))

	options = append(options, server.WithAddr(cfg.Host, cfg.Port))

	s, err := server.New(options...)
	if err != nil {
		log.Fatalf("[fh] server: %v\n", err)
	} else if err := s.Serve(context.TODO()); err != nil {
		log.Fatal(err)
	}

	return nil
}
