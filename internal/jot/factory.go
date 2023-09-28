// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

// NewFactory returns an initialized factory.
// Name and Path are for cookies.
// TTL is for all tokens.
func NewFactory(name, path string, ttl time.Duration) *Factory {
	if name == "" {
		name = "fh-jot"
	}
	if path == "" {
		path = "/"
	}
	f := Factory{
		ttl:     ttl,
		signers: make(map[string]Signer),
	}
	f.cookie.name = name
	f.cookie.path = path
	return &f
}

type Factory struct {
	sync.Mutex
	cookie struct {
		name string
		path string
	}
	ttl     time.Duration
	signers map[string]Signer
}

// AddSigner adds a new Signer.
func (f *Factory) AddSigner(signer Signer) error {
	if signer.Expired() {
		return ErrSignerExpired
	}
	f.Lock()
	defer f.Unlock()
	f.signers[signer.Id()] = signer

	return nil
}

// ClaimsFromToken extracts claims from a token.
// It returns an error if the token is invalid, expired, or hasn't been signed correctly.
// Otherwise, it returns the claims from the token.
func (f *Factory) ClaimsFromToken(token string) (*Claims, error) {
	// extract the header, claims, and signature from the token
	fields := strings.Split(token, ".")
	if len(fields) != 3 {
		return nil, ErrInvalidToken
	} else if len(fields[0]) > 99 {
		// header should be about 10 bytes for `typ`, 12 for `alg` and 40 for `kid`
		return nil, ErrInvalidHeader
	}
	// assign the fields to header, claims, and signature
	h64, c64, s64 := fields[0], fields[1], fields[2]

	// decode the header
	var header Header
	if data, err := decode_str(h64); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, &header); err != nil {
		return nil, err
	} else if header.TokenType != "JOT" {
		return nil, ErrUnknownType
	}

	// decode the signature
	signature, err := decode_str(s64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// use the header to find the original signer and confirm the signature
	signer, ok := f.signers[header.KeyID]
	if !ok {
		return nil, ErrInvalidSigner
	} else if signer.Algorithm() != header.Algorithm {
		return nil, ErrInvalidAlgorithm
	} else if !signer.Signed([]byte(h64+"."+c64), signature) {
		return nil, ErrInvalidSignature
	}

	// decode and validate the claims
	var claims Claims
	if data, err := decode_str(c64); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, &claims); err != nil {
		return nil, err
	} else if !claims.IsNotExpired(time.Now().UTC()) {
		return nil, ErrClaimsExpired
	}

	// return the claims if the message is signed and they haven't expired
	return &claims, nil
}

func (f *Factory) ClaimsToToken(ttl time.Duration, claims Claims) (string, error) {
	// select a signer at random
	signer, err := f.RandomSigner()
	if err != nil {
		return "", ErrMissingSigner
	}

	// update and marshal the header to JSON
	h64, err := Header{
		Algorithm: signer.Algorithm(),
		KeyID:     signer.Id(),
		TokenType: "JOT",
	}.Encode()
	if err != nil {
		return "", err
	}

	// update and marshal the claims to JSON
	iat := time.Now().UTC()
	c64, err := Claims{
		IssuedAt:  NumericDate(iat),
		ExpiresAt: NumericDate(iat.Add(ttl)),
		Roles:     claims.Roles,
		Subject:   claims.Subject,
	}.Encode()
	if err != nil {
		return "", err
	}

	// create message as header + '.' + claims
	var msg []byte
	msg = append(msg, h64...)
	msg = append(msg, '.')
	msg = append(msg, c64...)

	// sign the message
	signature, err := signer.Sign(msg)
	if err != nil {
		return "", err
	}
	s64 := encode_bytes(signature)

	// return the token as message + '.' + signature
	msg = append(msg, '.')
	msg = append(msg, s64...)
	return string(msg), nil
}

// DeleteSigner removes an existing Signer.
func (f *Factory) DeleteSigner(id string) {
	f.Lock()
	defer f.Unlock()
	delete(f.signers, id)
}

func (f *Factory) Lookup(id, algorithm string) (Signer, bool) {
	f.Lock()
	defer f.Unlock()

	signer, ok := f.signers[id]
	if !ok {
		return nil, false
	} else if signer.Expired() {
		delete(f.signers, id)
		return nil, false
	} else if signer.Algorithm() != algorithm {
		return nil, false
	}
	return signer, ok
}

func (f *Factory) PurgeExpiredSigners() {
	f.Lock()
	defer f.Unlock()

	for id, signer := range f.signers {
		if signer.Expired() {
			delete(f.signers, id)
		}
	}
}

func (f *Factory) RandomSigner() (Signer, error) {
	f.Lock()
	defer f.Unlock()

	for id, signer := range f.signers {
		if !signer.Expired() {
			return signer, nil
		}
		delete(f.signers, id)
	}
	return nil, ErrNotFound
}
