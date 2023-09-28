// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"time"
)

// Signer interface
type Signer interface {
	// Algorithm returns the name of the algorithm used by the signer.
	// The Factory will set the JOT header's "alg" field to this value when it is signed.
	// Example: "HS256"
	Algorithm() string

	// Expire clears the signer's expiration date.
	Expire()

	// Expired returns true if the Signer is expired.
	Expired() bool

	// Id is the unique identifier for this signer
	Id() string

	// Sign returns a slice containing the signature of the message.
	Sign(msg []byte) ([]byte, error)

	// Signed returns true if the msg was signed by this Signer.
	Signed(msg, signature []byte) bool
}

// HS256Signer implements a Signer using HMAC256.
type HS256Signer struct {
	id  string
	exp time.Time
	key []byte
}

func NewHS256Signer(id string, secret []byte, ttl time.Duration) (*HS256Signer, error) {
	return &HS256Signer{
		id:  id,
		key: append([]byte{}, secret...),
		exp: time.Now().Add(ttl).UTC(),
	}, nil
}

// Algorithm implements the Signer interface
func (s *HS256Signer) Algorithm() string {
	return "HS256"
}

// Expire implements the Signer interface.
func (s *HS256Signer) Expire() {
	s.exp = time.Unix(0, 0)
}

// Expired implements the Signer interface.
func (s *HS256Signer) Expired() bool {
	return s.exp.Before(time.Now().UTC())
}

// Id implements the Signer interface.
func (s *HS256Signer) Id() string {
	return s.id
}

// Sign implements the Signer interface.
func (s *HS256Signer) Sign(msg []byte) ([]byte, error) {
	hm := hmac.New(sha256.New, s.key)
	if _, err := hm.Write(msg); err != nil {
		return nil, err
	}
	return hm.Sum(nil), nil
}

// Signed implements the Signer interface.
func (s *HS256Signer) Signed(msg, signature []byte) bool {
	ours, err := s.Sign([]byte(msg))
	return err == nil && bytes.Equal(signature, ours)
}
