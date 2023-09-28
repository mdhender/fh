// Copyright (c) 2023 Michael D Henderson. All rights reserved.

package jot

import "encoding/base64"

// decode_bytes is a helper for base-64 decoding.
func decode_bytes(src []byte) ([]byte, error) {
	dst := make([]byte, base64.RawURLEncoding.DecodedLen(len(src)))
	_, err := base64.RawURLEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

// decode_str is a helper for base-64 decoding.
func decode_str(src string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(src)
}

// encode_bytes is a helper for base-64 encoding
func encode_bytes(src []byte) []byte {
	dst := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(dst, src)
	return dst
}

// encode_str is a helper for base-64 encoding
func encode_str(src string) []byte {
	dst := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(dst, []byte(src))
	return dst
}
