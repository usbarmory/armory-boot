// Copyright (c) The armory-boot authors. All Rights Reserved.
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

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
	"github.com/usbarmory/tamago/soc/nxp/usdhc"
)

func init() {
	log.SetFlags(0)

	switch imx6ul.Model() {
	case "i.MX6ULL", "i.MX6ULZ":
		imx6ul.SetARMFreq(imx6ul.FreqMax)
	case "i.MX6UL":
		imx6ul.SetARMFreq(imx6ul.Freq528)
	}
}

func preLaunch() {
	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)
}

func main() {
	var card *usdhc.USDHC

	usbarmory.LED("blue", false)
	usbarmory.LED("white", false)

	switch Boot {
	case "eMMC":
		card = usbarmory.MMC
	case "uSD":
		card = usbarmory.SD
	default:
		panic("invalid boot parameter")
	}

	part, err := disk.Detect(card, Start)

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

	log.Printf("armory-boot: starting kernel@%.8x\n", image.Entry())

	if err = image.Boot(preLaunch); err != nil {
		panic(fmt.Sprintf("load error, %v\n", err))
	}
}
