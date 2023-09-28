// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import (
	"encoding/json"
)

// Header is the header from a JOT.
type Header struct {
	Algorithm   string `json:"alg"`            // message authentication code algorithm, required
	ContentType string `json:"cty,omitempty"`  // not implemented
	Critical    string `json:"crit,omitempty"` // not implemented
	KeyID       string `json:"kid"`            // identifier used to sign, required
	TokenType   string `json:"typ"`            // should always be JOT, required
}

// Encode marshals the Header to JSON, then encodes the result as Base64.
func (h Header) Encode() ([]byte, error) {
	data, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	return encode_bytes(data), nil
}
