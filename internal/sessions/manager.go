// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
)

func (m *Manager) SessionFromRequest(r *http.Request) Session {
	token := tokenFromRequest(r, "fh-sess")
	if token == "" {
		return Session{}
	}
	ss, err := m.SessionFromToken(token)
	if err != nil {
		// log it?
		return Session{}
	}
	return ss
}

func (m *Manager) SessionFromToken(token string) (Session, error) {
	// extract the session id and signature from the token
	fields := strings.Split(token, ".")
	if len(fields) != 2 {
		return Session{}, ErrInvalidToken
	}
	id, sig64 := fields[0], fields[1]

	expectedSig64, err := m.Sign(id)
	if err != nil {
		return Session{}, err
	} else if sig64 != expectedSig64 {
		return Session{}, ErrInvalidSignature
	}

	ss, ok := m.Fetch(id)
	if !ok {
		return Session{}, ErrNotFound
	}

	return ss, nil
}

func (m *Manager) Sign(token string) (string, error) {
	hm := hmac.New(sha256.New, m.signing.key)
	if m.signing.salt != nil {
		if _, err := hm.Write(m.signing.salt); err != nil {
			return "", err
		}
	}
	if _, err := hm.Write([]byte(token)); err != nil {
		return "", err
	}
	if m.signing.pepper != nil {
		if _, err := hm.Write(m.signing.pepper); err != nil {
			return "", err
		}
	}
	return base64.RawURLEncoding.EncodeToString(hm.Sum(nil)), nil
}
