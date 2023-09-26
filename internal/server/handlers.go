// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
)

// getAssets returns a handler that serves static assets.
// If the path is a directory, it will serve path/index.html instead.
//
// Code based on http.ServeFile, but updated to refuse a directory listing.
func (s *Server) getAssets() http.HandlerFunc {
	root := s.assets.public
	log.Printf("[assets] path %s\n", root)

	// confirm that the root exists and is accessible.
	// note that changing the dir after this defeats the test.
	stat, err := os.Stat(root)
	if err != nil {
		log.Printf("[assets] %q: %+v\n", root, err)
		return func(w http.ResponseWriter, r *http.Request) {
			s.internalError(w, r, err)
		}
	} else if !stat.IsDir() {
		log.Printf("[assets] %q: must be a valid directory\n", root)
		return func(w http.ResponseWriter, r *http.Request) {
			s.internalError(w, r, fmt.Errorf("invalid path"))
		}
	}

	// return the handler that serves this all up.
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		// clean up the path in the url and trim the leading slash
		pathName := root + path.Clean("/"+r.URL.Path)

		stat, err := os.Stat(pathName)
		if err != nil {
			s.notFound(w, r)
			return
		} else if stat.IsDir() {
			// try to fetch path/index.html instead
			pathName = pathName + "/index.html"
			if stat, err = os.Stat(pathName); err != nil {
				// never give a directory listing
				s.notFound(w, r)
				return
			}
		}

		// path exists, but we only want to serve regular files
		if !stat.Mode().IsRegular() {
			s.notFound(w, r)
			return
		}

		file, err := os.Open(pathName)
		if err != nil {
			// pretty sure this should never happen!
			log.Printf("[assets] %q: %+v\n", pathName, err)
			s.notFound(w, r)
			return
		}
		// must close this file or leak resources
		defer func() {
			_ = file.Close()
		}()

		http.ServeContent(w, r, pathName, stat.ModTime(), file)
	}
}

func (s *Server) getIndex(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, nil, "index")
}

func (s *Server) getVersion(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Version string
	}{
		Version: s.version,
	}
	s.render(w, r, payload, "version")
}

func (s *Server) internalError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v\n", r.Method, r.URL, err)
	// http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

	payload := struct {
		Method string
		URL    string
		Error  error
	}{
		Method: r.Method,
		URL:    r.URL.Path,
		Error:  err,
	}
	s.render(w, r, payload, "internal_error")
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Method string
		URL    string
	}{
		Method: r.Method,
		URL:    r.URL.Path,
	}
	s.render(w, r, payload, "not_found")
}
