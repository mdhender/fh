// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

// Errors used by the package.
const (
	ErrBadFactory       = constError("bad factory")
	ErrBadSigner        = constError("bad signer")
	ErrClaimsExpired    = constError("claims expired")
	ErrInvalidAlgorithm = constError("invalid algorithm")
	ErrInvalidHeader    = constError("invalid header")
	ErrInvalidSignature = constError("invalid signature")
	ErrInvalidSigner    = constError("invalid signer")
	ErrInvalidToken     = constError("invalid token")
	ErrMissingClaims    = constError("missing claims")
	ErrMissingSigner    = constError("missing signer")
	ErrNotFound         = constError("not found")
	ErrSignerExpired    = constError("signer expired")
	ErrUnauthorized     = constError("unauthorized")
	ErrUnknownType      = constError("unknown type")
)

// declarations to support constant errors
type constError string

func (ce constError) Error() string {
	return string(ce)
}
