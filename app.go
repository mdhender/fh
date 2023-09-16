// Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package fh implements a game engine and web server for Far Horizons.
package fh

import (
	"crypto/tls"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mdhender/fh/pkg/mw"
	"github.com/mdhender/fh/pkg/semver"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type App struct {
	server  http.Server
	version semver.Version
}

func New(opts ...Option) (*App, error) {
	var options Options
	options.server.port = "8080"
	options.server.maxHeaderBytes = 1_048_576 // 1mb
	options.server.readTimeout = 5 * time.Second
	options.server.writeTimeout = 10 * time.Second

	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}

	if options.workingDir != "" {
		log.Printf("[app] changing working dir to %q\n", options.workingDir)
		if err := os.Chdir(options.workingDir); err != nil {
			return nil, fmt.Errorf("workingDir: %w", err)
		}
	}

	router := chi.NewRouter()
	if options.server.middleware.badRunes {
		router.Use(mw.BadRunes)
	}
	if options.server.middleware.cors {
		router.Use(mw.CORS)
	}
	if options.server.middleware.logging {
		router.Use(middleware.Logger)
	}

	a := &App{
		server: http.Server{
			Addr:           net.JoinHostPort(options.server.host, options.server.port),
			Handler:        router,
			MaxHeaderBytes: options.server.maxHeaderBytes,
			ReadTimeout:    options.server.readTimeout,
			WriteTimeout:   options.server.writeTimeout,
		},
		version: semver.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		},
	}
	a.Routes(router)

	return a, nil
}

func (a *App) ListenAndServe() error {
	return a.server.ListenAndServe()
}

// ListenAndServeTLS jams in an untested TLS config.
// Should probably validate against notes from
// https://github.com/denji/golang-tls
// and https://eli.thegreenplace.net/2021/go-https-servers-with-tls/.
// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
func (a *App) ListenAndServeTLS(certFile, keyFile string) error {
	a.server.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	a.server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))

	return a.server.ListenAndServeTLS(certFile, keyFile)
}
