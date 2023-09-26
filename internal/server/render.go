// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// render loads templates, renders the data, and writes it to the response.
func (s *Server) render(w http.ResponseWriter, r *http.Request, data any, names ...string) {
	files := []string{filepath.Join(s.assets.templates, "layout.gohtml")}
	for _, name := range names {
		files = append(files, filepath.Join(s.assets.templates, name+".gohtml"))
	}
	files = append(files, filepath.Join(s.assets.templates, "navbar.gohtml"))

	w.Header().Set("FH-Version", s.version)

	t, err := template.ParseFiles(files...)
	if err != nil {
		log.Printf("%s %s: render: parse: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("%s %s: render: execute: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
