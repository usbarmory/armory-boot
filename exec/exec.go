// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

// BootImage represents a bootable image.
type BootImage interface {
	Load() error
	Boot(cleanup func()) error
}
