// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

// BootImage represents a bootable image.
type BootImage interface {
	Load() error
	Entry() uint
	Boot(cleanup func()) error
}
