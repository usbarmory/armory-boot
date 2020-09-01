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

	"github.com/jedisct1/go-minisign"
)

const signatureSuffix = ".sig"

var PublicKey string

func verifySignature(bin []byte, s []byte) (valid bool, err error) {
	sig, err := minisign.DecodeSignature(string(s))

	if err != nil {
		return false, fmt.Errorf("invalid signature, %v\n", err)
	}

	pub, err := minisign.NewPublicKey(PublicKey)

	if err != nil {
		return false, fmt.Errorf("invalid public key, %v\n", err)
	}

	return pub.Verify(bin, sig)
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
