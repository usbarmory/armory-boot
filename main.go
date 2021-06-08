// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/f-secure-foundry/armory-boot/config"
	"github.com/f-secure-foundry/armory-boot/disk"
	"github.com/f-secure-foundry/armory-boot/exec"

	"github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/f-secure-foundry/tamago/soc/imx6"
	"github.com/f-secure-foundry/tamago/soc/imx6/rngb"
)

var Build string
var Revision string

var Boot string
var Start string

// signify/minisign authentication when defined
var PublicKeyStr string

func init() {
	log.SetFlags(0)

	if err := imx6.SetARMFreq(900); err != nil {
		panic(fmt.Sprintf("cannot change ARM frequency, %v\n", err))
	}
}

func initBootMedia(device string, start string) (part *disk.Partition, err error) {
	offset, err := strconv.ParseInt(start, 10, 64)

	if err != nil {
		return nil, fmt.Errorf("invalid start offset, %v\n", err)
	}

	part = &disk.Partition{
		Offset: offset,
	}

	switch device {
	case "eMMC":
		part.Card = usbarmory.MMC
	case "uSD":
		part.Card = usbarmory.SD
	default:
		return nil, errors.New("invalid boot parameter")
	}

	if err := part.Card.Detect(); err != nil {
		return nil, fmt.Errorf("boot media error, %v\n", err)
	}

	return
}

func preLaunch() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	// RNGB driver doesn't play well with previous initializations
	rngb.Reset()
}

func main() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	part, err := initBootMedia(Boot, Start)

	if err != nil {
		panic(fmt.Sprintf("boot media error, %v\n", err))
	}

	usbarmory.LED("blue", true)

	if len(PublicKeyStr) == 0 {
		log.Printf("armory-boot: no public key, skipping signature verification")
	}

	conf, err := config.Load(part, config.DefaultConfigPath, config.DefaultSignaturePath, PublicKeyStr)

	if err != nil {
		panic(fmt.Sprintf("configuration error, %v\n", err))
	}

	log.Printf("\n%s", conf.JSON)

	usbarmory.LED("white", true)

	if conf.ELF {
		err = exec.BootELF(mem, conf.Kernel(), preLaunch)
	} else {
		err = exec.BootLinux(mem, &exec.LinuxImage{
			Kernel:               conf.Kernel(),
			DeviceTreeBlob:       conf.DeviceTreeBlob(),
			InitialRamDisk:       conf.InitialRamDisk(),
			KernelOffset:         kernelOffset,
			DeviceTreeBlobOffset: paramsOffset,
			InitialRamDiskOffset: initrdOffset,
			CmdLine:              conf.CmdLine,
		}, preLaunch)
	}

	if err != nil {
		panic(fmt.Sprintf("boot error, %v\n", err))
	}
}
