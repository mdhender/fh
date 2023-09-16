// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package main

import (
	"github.com/mdhender/fh/pkg/homedir"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)

	defer func(started time.Time) {
		if args.time {
			log.Printf("elapsed time %v\n", time.Now().Sub(started))
		}
	}(time.Now())

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	args.home, err = homedir.Dir()
	if err != nil {
		return err
	}

	cmdRoot.PersistentFlags().BoolVar(&args.time, "time", args.time, "display run time statistics on completion")

	cmdRoot.AddCommand(cmdServe)
	cmdServe.Flags().BoolVar(&args.middleware.badRunes, "bad-runes-middleware", args.middleware.badRunes, "enable bad runes middleware")
	cmdServe.Flags().BoolVar(&args.middleware.cors, "cors-middleware", args.middleware.cors, "enable CORS options middleware")
	cmdServe.Flags().BoolVar(&args.middleware.logging, "logging-middleware", args.middleware.logging, "enable logging middleware")
	cmdServe.Flags().StringVar(&args.host, "host", args.templates, "host to bind to")
	cmdServe.Flags().StringVar(&args.port, "port", args.templates, "port to listen on")
	cmdServe.Flags().StringVar(&args.templates, "templates", args.templates, "path to templates")
	cmdServe.Flags().StringVar(&args.workingDir, "working-dir", args.workingDir, "path to execute in")

	return cmdRoot.Execute()
}

// args is the global arguments
var args struct {
	home       string
	host       string
	middleware struct {
		badRunes bool
		cors     bool
		logging  bool
	}
	port       string
	templates  string
	time       bool
	workingDir string
}

// cmdRoot represents the base command when called without any subcommands
var cmdRoot = &cobra.Command{
	Use:   "fh",
	Short: "Far Horizons server",
	Long: `This application implements a game engine for Far Horizons
and a web server for players.`,
}
