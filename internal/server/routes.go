// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

import (
	"net/http"
)

func (s *Server) Routes() {
	s.router.HandleFunc("GET", "/", s.getIndex)
	s.router.HandleFunc("GET", "/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	s.router.HandleFunc("GET", "/signout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Path:     "/",
			Name:     "fh-auth",
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	s.router.HandleFunc("POST", "/signout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Path:     "/",
			Name:     "fh-auth",
			HttpOnly: true,
			Secure:   true,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	s.router.HandleFunc("GET", "/version", s.getVersion)

	// assume that anything else is a static asset and attempt to serve it.
	s.router.NotFound = s.getAssets()
}
