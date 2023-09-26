// Copyright (c) 2023 Michael D Henderson. All rights reserved.

// Package server implements a web server for Far Horizons.
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/mdhender/fh/internal/semver"
	"github.com/mdhender/fh/internal/way"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	server http.Server
	assets struct {
		public    string
		templates string
	}
	context context.Context
	do      struct {
		log bool
	}
	router *way.Router
	tls    struct {
		enabled  bool
		certFile string
		keyFile  string
	}
	version string
}

func New(options ...Option) (*Server, error) {
	s := &Server{
		context: context.TODO(),
		router:  way.NewRouter(),
		version: semver.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		}.String(),
	}
	s.server.Addr = net.JoinHostPort("", "3000")
	s.server.Handler = s.router
	s.server.MaxHeaderBytes = 1_048_576 // 1mb
	s.server.IdleTimeout = 30 * time.Second
	s.server.ReadTimeout = 5 * time.Second
	s.server.WriteTimeout = 10 * time.Second

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	s.Routes()

	log.Printf("[serve] public    %s\n", s.assets.public)
	log.Printf("[serve] templates %s\n", s.assets.templates)

	return s, nil
}

// Run starts the embedded http.Server so that it will gracefully handle receipt of SIGTERM or SIGINT.
// From https://clavinjune.dev/en/blogs/golang-http-server-graceful-shutdown/.
func (s *Server) Run() error {
	// start the server in a new go routine
	go func(ctx context.Context) {
		log.Printf("[server] listening on %s\n", s.server.Addr)
		if err := s.Serve(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
		log.Printf("[server] server stopped gracefully\n")
	}(s.context)

	// create channels to catch signals
	stopCh, closeCh := s.SignalChannels()
	defer func() {
		closeCh()
	}()
	log.Println("[server] stopCh: notified: ", <-stopCh)

	return nil
}

// Serve is a wrapper around ListenAndServe and ListenAndServeTLS.
func (s *Server) Serve(ctx context.Context) (err error) {
	s.context = ctx
	if s.tls.enabled {
		// ListenAndServeTLS jams in an untested TLS config.
		// Should probably validate against notes from
		// https://github.com/denji/golang-tls
		// and https://eli.thegreenplace.net/2021/go-https-servers-with-tls/.
		// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
		s.server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
		s.server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
		err = s.server.ListenAndServeTLS(s.tls.certFile, s.tls.keyFile)
	} else {
		err = s.server.ListenAndServe()
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// SignalChannels creates channels to catch signals.
// From https://clavinjune.dev/en/blogs/golang-http-server-graceful-shutdown/.
func (s *Server) SignalChannels() (chan os.Signal, func()) {
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	return stopCh, func() {
		close(stopCh)
	}
}

// Shutdown catches signals and attempts a graceful shutdown.
// From https://clavinjune.dev/en/blogs/golang-http-server-graceful-shutdown/.
func (s *Server) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		panic(err)
	}
	log.Printf("[server] caught shutdown signal\n")
}
