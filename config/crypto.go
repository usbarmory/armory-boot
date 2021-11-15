// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package config

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/dcp"
)

func init() {
	if imx6.Native {
		dcp.Init()
	}
}

// Verify authenticates an input against a signify/minisign generated
// signature, pubKey must be the last line of a signify/minisign public key
// (i.e. without comments).
func Verify(buf []byte, sig []byte, pubKey string) (err error) {
	s, err := DecodeSignature(string(sig))

	if err != nil {
		return fmt.Errorf("invalid signature, %v", err)
	}

	pub, err := NewPublicKey(pubKey)

	if err != nil {
		return fmt.Errorf("invalid public key, %v", err)
	}

	valid, err := pub.Verify(buf, s)

	if err != nil {
		return fmt.Errorf("invalid signature, %v", err)
	}

	if !valid {
		return errors.New("invalid signature")
	}

	return
}

// CompareHash computes a SHA256 checksum of the input data, using hardware
// acceleration (NXP DCP), and compares the computed hash with the one passed
// as a string with only hexadecimal characters and even length.
//
// This function is only meant to be used with `GOOS=tamago GOARCH=arm` on
// i.MX6 targets.
func CompareHash(buf []byte, s string) (valid bool) {
	sum, err := dcp.Sum256(buf)

	if err != nil {
		return false
	}

	hash, err := hex.DecodeString(s)

	if err != nil {
		return false
	}

	return bytes.Equal(sum[:], hash)
}
