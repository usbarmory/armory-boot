// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package exec provides support for kernel image loading and booting in bare
// metal Go applications.
//
// This package is only meant to be used with `GOOS=tamagom`
// as supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package exec

// BootImage represents a bootable image.
type BootImage interface {
	Load() error
	Entry() uint
	Boot(cleanup func()) error
}
