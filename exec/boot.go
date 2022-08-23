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
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package exec

import (
	"errors"
	"log"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

// defined in boot.s
func exec(kernel uint32, params uint32)
func svc()

func boot(kernel uint32, params uint32, cleanup func()) (err error) {
	arm.SystemExceptionHandler = func(n int) {
		if n != arm.SUPERVISOR {
			panic("unhandled exception")
		}

		cleanup()

		imx6ul.ARM.DisableInterrupts()
		imx6ul.ARM.FlushDataCache()
		imx6ul.ARM.DisableCache()

		log.Printf("armory-boot: starting kernel@%x params@%x\n", kernel, params)
		exec(kernel, params)
	}

	svc()

	return errors.New("supervisor failure")
}
