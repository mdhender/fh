// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/fh"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

// cmdServe implements the command to start the application server.
var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "start application server",
	Long:  `This command starts the application server for Far Horizons.`,
	Run: func(cmd *cobra.Command, _ []string) {
		var options []fh.Option
		var err error

		if args.templates, err = filepath.Abs(args.templates); err != nil {
			log.Fatalf("[app] templates: %v\n", err)
		} else if sb, err := os.Stat(args.templates); err != nil {
			log.Fatalf("[app] templates: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("[app] templates: %v\n", fmt.Errorf("invalid path %q", args.templates))
		}
		options = append(options, fh.WithTemplates(args.templates))

		if args.workingDir != "" {
			if args.workingDir, err = filepath.Abs(args.workingDir); err != nil {
				log.Fatalf("[app] workingDir: %v\n", err)
			} else if sb, err := os.Stat(args.workingDir); err != nil {
				log.Fatalf("[app] workingDir: %v\n", err)
			} else if !sb.IsDir() {
				log.Fatalf("[app] workingDir: %v\n", fmt.Errorf("invalid path %q", args.workingDir))
			}
			options = append(options, fh.WithWorkingDir(args.workingDir))
		}

		options = append(options, fh.WithBadRunesMiddleware(args.middleware.badRunes))
		options = append(options, fh.WithCorsMiddleware(args.middleware.cors))
		options = append(options, fh.WithLoggingMiddleware(args.middleware.logging))

		app, err := fh.New(options...)
		if err != nil {
			log.Fatalf("[contacts] app: %v\n", err)
		}

		if err := app.ListenAndServe(); err != nil {
			log.Fatal(err)
		}

	},
}
