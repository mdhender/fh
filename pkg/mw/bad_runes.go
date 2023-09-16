// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package mw

import (
	"log"
	"net/http"
	"unicode"
)

// BadRunes will return an error if the URL contains any non-printable runes.
func BadRunes(next http.Handler) http.Handler {
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
