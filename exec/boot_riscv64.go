// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.
package exec

import (
	"github.com/usbarmory/tamago/dma"
)

func boot(kernel uint, params uint, cleanup func(), _ *dma.Region) (err error) {
	panic("unimplemented")
}
