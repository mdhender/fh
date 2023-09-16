// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package fh

import (
	"fmt"
	"github.com/go-chi/chi"
	"net/http"
)

func (a *App) Routes(r *chi.Mux) {
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "<p>Far Horizons v%s\n</p>", a.version.String())
	})
}
