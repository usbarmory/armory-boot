// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "unsafe"
)

// Override imx6ul.ramStart, usbarmory.ramSize and dma allocation, as the
// mapped kernel image needs to be within the first 128MiB of RAM.

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = 0x90000000

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = 0x10000000

const (
	dmaStart = 0x80000000
	dmaSize  = 0x10000000
)

const (
	kernelOffset = 0x0800000
	paramsOffset = 0x7000000
	initrdOffset = 0x8000000
)
