// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"log"
	"net/http"
	"strings"
)

// FromBearerToken extracts and returns a bearer token from the request.
// Returns an empty string if there is no bearer token or the token is invalid.
func FromBearerToken(r http.Request) string {
	// first try a bearer token
	log.Printf("jot: bearer: entered\n")
	headerAuthText := r.Header.Get("Authorization")
	if headerAuthText == "" {
		return ""
	}
	log.Printf("jot: bearer: found authorization header\n")
	authTokens := strings.SplitN(headerAuthText, " ", 2)
	if len(authTokens) != 2 {
		return ""
	}
	log.Printf("jot: bearer: found authorization token\n")
	authType, authToken := authTokens[0], strings.TrimSpace(authTokens[1])
	if authType != "Bearer" {
		return ""
	}
	log.Printf("jot: bearer: found bearer token\n")
	return authToken
}

// FromCookie extracts and returns a token from a cookie in the request.
// Returns an empty string if there is no cookie or the token is invalid.
func FromCookie(r http.Request, cookie string) string {
	log.Printf("jot: cookie: entered\n")
	c, err := r.Cookie(cookie)
	if err != nil {
		log.Printf("jot: cookie: %+v\n", err)
		return ""
	}
	log.Printf("jot: cookie: token\n")
	return c.Value
}

// FromRequest extracts a token from a request.
// It tries to find a bearer token first.
// If it can't, it searches for a cookie.
func FromRequest(r http.Request, cookie string) string {
	token := FromBearerToken(r)
	if token == "" {
		token = FromCookie(r, cookie)
	}
	return token
}
