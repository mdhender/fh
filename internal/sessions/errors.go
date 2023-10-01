// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package sessions

// Errors used by the package.
const (
	ErrInvalidSignature = constError("invalid signature")
	ErrInvalidToken     = constError("invalid token")
	ErrNotFound         = constError("not found")
)

// declarations to support constant errors
type constError string

func (ce constError) Error() string {
	return string(ce)
}
