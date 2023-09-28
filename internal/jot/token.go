// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"time"
)

// JOT implements my version of the JSON Web JOT.
// It's like a JWT that's been customized for this application.
type JOT struct {
	Header    Header
	Claims    Claims
	Signature []byte
	isSigned  bool
}

// IsNotExpired returns true if the token has not expired.
func (j *JOT) IsNotExpired() bool {
	return j.Claims.IsNotExpired(time.Now().UTC())
}

// IsSigned returns true only if the signature has been verified.
func (j *JOT) IsSigned() bool {
	panic("!implemented")
}

// IsValid returns true only if the token is signed and not expired.
func (j *JOT) IsValid() bool {
	return !j.IsNotExpired() && j.IsSigned()
}

// String implements the Stringer interface.
func (j *JOT) String() string {
	panic("!implemented")
}
