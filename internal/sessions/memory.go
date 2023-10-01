// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Store is a session store.
type Store struct {
	sync.Mutex
	sessions map[string]Session
	signing  struct {
		salt   []byte
		key    []byte
		pepper []byte
		ttl    time.Duration
	}
}

func (s *Store) Delete(id string) {
	s.Lock()
	defer s.Unlock()
	delete(s.sessions, id)
}

func (s *Store) FetchFromRequest(r *http.Request) Session {
	// fetch the token from the request (either a cookie or bearer token)
	token := tokenFromRequest(r, "fh-sess")
	if token == "" {
		return Session{}
	}
	// extract the session id and signature from the token
	fields := strings.Split(token, ".")
	if len(fields) != 2 {
		return Session{}
	}
	id, sig64 := fields[0], fields[1]
	// determine the expected signature
	expectedSig64, err := s.Sign(id)
	if err != nil {
		return Session{}
	}
	// and compare the two
	if sig64 != expectedSig64 {
		return Session{}
	}
	// token is valid, so return the associated session
	sess, ok := s.Lookup(id)
	if !ok {
		return Session{}
	}
	return sess
}

func (s *Store) Lookup(id string) (Session, bool) {
	s.Lock()
	defer s.Unlock()

	sess, ok := s.sessions[id]
	if !ok {
		return Session{}, false
	} else if sess.IsExpired() {
		delete(s.sessions, id)
	}
	return Session{}, false
}

// Sign returns a Base64 encoding of the token's hash.
func (s *Store) Sign(token string) (string, error) {
	hm := hmac.New(sha256.New, s.signing.key)
	if s.signing.salt != nil {
		if _, err := hm.Write(s.signing.salt); err != nil {
			return "", err
		}
	}
	if _, err := hm.Write([]byte(token)); err != nil {
		return "", err
	}
	if s.signing.pepper != nil {
		if _, err := hm.Write(s.signing.pepper); err != nil {
			return "", err
		}
	}
	return base64.RawURLEncoding.EncodeToString(hm.Sum(nil)), nil
}
