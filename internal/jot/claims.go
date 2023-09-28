// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"encoding/json"
	"time"
)

// Claims is sometimes called the payload.
type Claims struct {
	// The recipients that the JOT is intended for.
	// Each principal intended to process the JOT must identify itself with a value in the audience claim.
	// If the principal processing the claim does not identify itself with a value in the aud claim when this claim is present,
	// then the JOT must be rejected.
	// Not implemented.
	// Audience []string `json:"aud,omitempty"`

	// The expiration time on and after which the JOT must not be accepted for processing.
	ExpiresAt NumericDate `json:"exp"`

	// The time at which the JOT was issued.
	// Not implemented.
	IssuedAt NumericDate `json:"iat,omitempty"`

	// The principal that issued the JOT.
	// Not implemented.
	// Issuer string `json:"iss,omitempty"`

	// Case-sensitive unique identifier of the token even among different issuers.
	// Not implemented.
	// JWTID string `json:"jti,omitempty"`

	// The time on which the JOT will start to be accepted for processing.
	// Not implemented.
	// NotBefore NumericDate `json:"nbf,omitempty"`

	// The roles assigned to the subject.
	Roles Roles `json:"roles"`

	// The subject of the JOT.
	Subject string `json:"sub"`
}

// Encode marshals the Claims to JSON, then encodes the result as Base64.
func (c Claims) Encode() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return encode_bytes(data), nil
}

// IsExpired returns true if `expiresAt` is not after `now`.
func (c Claims) IsExpired(now time.Time) bool {
	return !time.Time(c.ExpiresAt).After(now)
}

// IsNotExpired returns true if `expiresAt` is after `now`.
func (c Claims) IsNotExpired(now time.Time) bool {
	return time.Time(c.ExpiresAt).After(now)
}
