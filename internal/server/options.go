// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Options struct {
	addr           string
	host           string
	maxHeaderBytes int
	middleware     []func(http.Handler) http.Handler
	port           string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	templates      string // absolute path to templates directory
	workingDir     string // absolute path to working directory
}

type Option func(*Server) error

func WithAddr(host, port string) Option {
	return func(s *Server) (err error) {
		s.server.Addr = net.JoinHostPort(host, port)
		return nil
	}
}

func WithAssets(kind, path string) Option {
	return func(s *Server) error {
		if sb, err := os.Stat(path); err != nil {
			return fmt.Errorf("assets %q: %w", kind, err)
		} else if !sb.IsDir() {
			return fmt.Errorf("assets %q: %w", kind, fmt.Errorf("not a directory"))
		} else if path, err = filepath.Abs(path); err != nil {
			return fmt.Errorf("assets %q: %w", kind, err)
		}
		switch kind {
		case "public":
			s.assets.public = path
		case "templates":
			s.assets.templates = path
		default:
			return fmt.Errorf("unknown asset %q", kind)
		}
		return nil
	}
}

func WithMaxBodyLength(l int) Option {
	return func(s *Server) (err error) {
		s.server.MaxHeaderBytes = l
		return nil
	}
}

func WithMiddleware(mw func(http.Handler) http.Handler) Option {
	return func(s *Server) error {
		return fmt.Errorf("not implemented")
	}
}

func WithTLS(cert, key string) Option {
	return func(s *Server) error {
		if sb, err := os.Stat(cert); err != nil {
			return fmt.Errorf("tls cert: %w", err)
		} else if mode := sb.Mode(); mode.IsDir() || !mode.IsRegular() {
			return fmt.Errorf("tls cert: %w", fmt.Errorf("not a regular file"))
		} else if cert, err = filepath.Abs(cert); err != nil {
			return fmt.Errorf("tls cert: %w", err)
		}
		s.tls.certFile = cert

		if sb, err := os.Stat(key); err != nil {
			return fmt.Errorf("tls key: %w", err)
		} else if mode := sb.Mode(); mode.IsDir() || !mode.IsRegular() {
			return fmt.Errorf("tls key: %w", fmt.Errorf("not a regular file"))
		} else if cert, err = filepath.Abs(cert); err != nil {
			return fmt.Errorf("tls key: %w", err)
		}
		s.tls.keyFile = key

		s.tls.enabled = true

		return nil
	}
}

func WithTimeout(kind string, ttl time.Duration) Option {
	return func(s *Server) error {
		switch kind {
		case "idle":
			s.server.IdleTimeout = ttl
		case "read":
			s.server.ReadTimeout = ttl
		case "write":
			s.server.WriteTimeout = ttl
		default:
			return fmt.Errorf("unknown timeout %q", kind)
		}
		return nil
	}
}
