// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/f-secure-foundry/tamago/imx6"
	"github.com/f-secure-foundry/tamago/imx6/usdhc"
	"github.com/f-secure-foundry/tamago/usbarmory/mark-two"
)

var Build string
var Revision string

var Boot string
var Start string

func init() {
	log.SetFlags(0)

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("WARNING: error setting ARM frequency: %v\n", err))
	}
}

func main() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	var card *usdhc.USDHC

	switch Boot {
	case "eMMC":
		card = usbarmory.MMC
	case "uSD":
		card = usbarmory.SD
	default:
		panic("invalid boot parameter")
	}

	if err := card.Detect(); err != nil {
		panic(err)
	}

	usbarmory.LED("white", true)

	offset, err := strconv.ParseInt(Start, 10, 64)

	if err != nil {
		panic(err)
	}

	partition := &Partition{
		Card:   card,
		Offset: offset,
	}

	err = conf.Read(partition, defaultConfigPath)

	if err != nil {
		panic(err)
	}

	kernel, err := partition.ReadAll(conf.Kernel[0])

	if err != nil {
		panic(err)
	}

	dtb, err := partition.ReadAll(conf.DeviceTreeBlob[0])

	if err != nil {
		panic(err)
	}

	usbarmory.LED("blue", true)

	log.Printf("armory-boot: kernel %s (%d bytes)\n", conf.Kernel[0], len(kernel))
	log.Printf("armory-boot: dtb %s (%d bytes)\n", conf.DeviceTreeBlob[0], len(dtb))
	log.Printf("armory-boot: cmdline %s\n", conf.CmdLine)

	log.Printf("armory-boot: verifying kernel %s", conf.Kernel[1])

	if !verify(kernel, conf.Kernel[1]) {
		panic("invalid kernel hash")
	}

	log.Printf("armory-boot: verifying dtb %s", conf.DeviceTreeBlob[1])

	if !verify(dtb, conf.DeviceTreeBlob[1]) {
		panic("invalid dtb hash")
	}

	dtb, err = fixupDeviceTree(dtb, conf.CmdLine)

	if err != nil {
		panic(err)
	}

	// TODO: verify configuration signature

	boot(kernel, dtb, conf.CmdLine)
}
