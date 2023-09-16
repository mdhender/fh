// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package fh

import (
	"fmt"
	"os"
	"time"
)

type Options struct {
	server struct {
		host           string
		port           string
		addr           string
		maxHeaderBytes int
		readTimeout    time.Duration
		writeTimeout   time.Duration
		middleware     struct {
			badRunes bool
			cors     bool
			logging  bool
		}
	}
	templates  string // absolute path to templates directory
	workingDir string // absolute path to working directory
}

type Option func(*Options) error

func WithBadRunesMiddleware(use bool) Option {
	return func(o *Options) error {
		o.server.middleware.badRunes = use
		return nil
	}
}

func WithCorsMiddleware(use bool) Option {
	return func(o *Options) error {
		o.server.middleware.cors = use
		return nil
	}
}

func WithHost(host string) Option {
	return func(o *Options) (err error) {
		o.server.host = host
		return nil
	}
}

func WithLoggingMiddleware(use bool) Option {
	return func(o *Options) error {
		o.server.middleware.logging = use
		return nil
	}
}

func WithMaxBodyLength(l int) Option {
	return func(o *Options) (err error) {
		o.server.maxHeaderBytes = l
		return nil
	}
}
func WithPort(port string) Option {
	return func(o *Options) (err error) {
		o.server.port = port
		return nil
	}
}

func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) error {
		o.server.readTimeout = d
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(o *Options) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("invalid path %q", path)
		}
		o.templates = path
		return nil
	}
}

func WithWorkingDir(path string) Option {
	return func(o *Options) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("invalid path %q", path)
		}
		o.workingDir = path
		return nil
	}
}

func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) error {
		o.server.writeTimeout = d
		return nil
	}
}
