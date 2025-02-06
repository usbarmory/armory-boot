// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"errors"

	"github.com/usbarmory/tamago/dma"
)

// defined in boot_amd64.s
func exec(kernel uint, params uint)

func boot(kernel uint, params uint, cleanup func(), _ *dma.Region) (err error) {
	if cleanup != nil {
		cleanup()
	}

	exec(kernel, params)

	return errors.New("boot failure")
}
