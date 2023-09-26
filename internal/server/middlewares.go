// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

import (
	"log"
	"net/http"
	"strings"
	"unicode"
)

// BadRunes will return an error if the URL contains any non-printable runes.
func (s *Server) BadRunes(next http.Handler) http.Handler {
	log.Printf("[middleware] adding check for bad runes in request URL\n")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("[middleware] bad runes check\n")
		for _, ch := range r.URL.Path {
			if !unicode.IsPrint(ch) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// CORS will inject CORS headers on any OPTIONS request
func (s *Server) CORS(next http.Handler) http.Handler {
	log.Printf("[middleware] adding cors pre-flight middleware\n")

	allowHeaders := strings.Join([]string{
		"Accept",
		"Accept-Encoding",
		"Authorization",
		"Content-Length",
		"Content-Type",
		"X-CSRF-Token",
	}, ", ")
	allowMethods := strings.Join([]string{
		"DELETE",
		"GET",
		"HEAD",
		"OPTIONS",
		"POST",
		"PUT",
	}, ", ")
	maxAge := "300" // max age not ignored by any of major browsers

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// inject CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", allowMethods)
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		w.Header().Set("Access-Control-Max-Age", maxAge)

		// if we get the pre-flight request, return immediately
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logger will log a request.
func (s *Server) Logger(next http.Handler) http.Handler {
	log.Printf("[middleware] adding logger\n")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
