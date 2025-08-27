// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"errors"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// defined in boot_arm.s
func svc()
func exec()

// exec() parameters are passed as pointers to limit stack allocation as it's
// executed on g0
var (
	_kernel uint32
	_params uint32
	_mmu    bool
)

func boot(kernel uint, params uint, cleanup func(), region *dma.Region) (err error) {
	table := arm.SystemVectorTable()
	table.Supervisor = exec

	imx6ul.ARM.SetVectorTable(table)

	_kernel = uint32(kernel)
	_params = uint32(params)
	_mmu = (region != nil)

	if cleanup != nil {
		cleanup()
	}

	if region != nil {
		imx6ul.ARM.SetAttribute(
			uint32(region.Start()),
			uint32(region.End()),
			arm.TTE_EXECUTE_NEVER, 0)
	} else {
		imx6ul.ARM.FlushDataCache()
		imx6ul.ARM.DisableCache()
	}

	svc()

	return errors.New("supervisor failure")
}
