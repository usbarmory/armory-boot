// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package exec provides support for kernel image loading and booting on the
// USB armory Mk II platform.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package exec

import (
	"errors"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// defined in boot.s
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

	cleanup()

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
