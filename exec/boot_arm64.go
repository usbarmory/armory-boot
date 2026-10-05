// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/dma"
)

const requiredEL = 1

// defined in boot_arm64.s
func exec(kernel uint, params uint)

func boot(kernel uint, params uint, cleanup func(), _ *dma.Region) (err error) {
	cpu := arm64.CPU{}

	if el := cpu.CurrentEL(); el != requiredEL {
		return fmt.Errorf("EL%d is required, running at EL%d", requiredEL, el)
	}

	if cleanup != nil {
		cleanup()
	}

	cpu.FlushDataCache()
	cpu.DisableCache()

	exec(kernel, params)

	return errors.New("exec failure")
}
