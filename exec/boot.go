// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package exec provides support for kernel image loading and booting on the
// USB armory Mk II platform.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package exec

import (
	"errors"
	"log"

	"github.com/f-secure-foundry/tamago/arm"
	"github.com/f-secure-foundry/tamago/soc/imx6"
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

		imx6.ARM.DisableInterrupts()
		imx6.ARM.FlushDataCache()
		imx6.ARM.DisableCache()

		log.Printf("armory-boot: starting kernel@%x params@%x\n", kernel, params)
		exec(kernel, params)
	}

	svc()

	return errors.New("supervisor failure")
}
