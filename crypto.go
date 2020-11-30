// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func verifySignature(bin []byte, sig []byte, pubKey string) (valid bool, err error) {
	s, err := DecodeSignature(string(sig))

	if err != nil {
		return false, fmt.Errorf("invalid signature, %v", err)
	}

	pub, err := NewPublicKey(pubKey)

	if err != nil {
		return false, fmt.Errorf("invalid public key, %v", err)
	}

	return pub.Verify(bin, s)
}

func verifyHash(bin []byte, s string) bool {
	h := sha256.New()
	h.Write(bin)

	hash, err := hex.DecodeString(s)

	if err != nil {
		return false
	}

	return bytes.Equal(h.Sum(nil), hash)
}
