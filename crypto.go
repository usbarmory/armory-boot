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
)

func verify(buf []byte, s string) bool {
	h := sha256.New()
	h.Write(buf)

	hash, err := hex.DecodeString(s)

	if err != nil {
		return false
	}

	return bytes.Equal(h.Sum(nil), hash)
}
