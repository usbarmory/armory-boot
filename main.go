// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/usbarmory/armory-boot/config"
	"github.com/usbarmory/armory-boot/disk"
	"github.com/usbarmory/armory-boot/exec"

	"github.com/usbarmory/tamago/board/f-secure/usbarmory/mark-two"
	"github.com/usbarmory/tamago/soc/imx6"
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

func preLaunch() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)
}

func main() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	part, err := disk.Detect(Boot, Start)

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

	var image exec.BootImage

	if conf.ELF {
		image = &exec.ELFImage{
			Region: mem,
			ELF:    conf.Kernel(),
		}
	} else {
		image = &exec.LinuxImage{
			Region:               mem,
			Kernel:               conf.Kernel(),
			DeviceTreeBlob:       conf.DeviceTreeBlob(),
			InitialRamDisk:       conf.InitialRamDisk(),
			KernelOffset:         kernelOffset,
			DeviceTreeBlobOffset: paramsOffset,
			InitialRamDiskOffset: initrdOffset,
			CmdLine:              conf.CmdLine,
		}
	}

	if err = image.Load(); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}

	if err = image.Boot(preLaunch); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}
}
