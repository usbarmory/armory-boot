// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/f-secure-foundry/tamago/dma"

	"github.com/u-root/u-root/pkg/dt"
)

// LinuxImage represents a bootable Linux kernel image.
type LinuxImage struct {
	// Region is the memory area for image loading.
	Region *dma.Region

	// Kernel is the Linux kernel image.
	Kernel []byte
	// KernelOffset is the Linux kernel offset from RAM start address.
	KernelOffset int

	// DeviceTreeBlob is the Linux kernel dtb file.
	DeviceTreeBlob []byte
	// DeviceTreeBlobOffset is the dtb offset from RAM start address.
	DeviceTreeBlobOffset int

	// InitialRamDisk is the Linux kernel initrd file.
	InitialRamDisk []byte
	// InitialRamDiskOffset is the initrd offset from RAM start address.
	InitialRamDiskOffset int

	// CmdLine is the Linux kernel command line arguments.
	CmdLine string

	entry  uint32
	dtb    uint32
	loaded bool
}

func (image *LinuxImage) fdt() (fdt *dt.FDT, err error) {
	return dt.ReadFDT(bytes.NewReader(image.DeviceTreeBlob))
}

func (image *LinuxImage) updateDTB(fdt *dt.FDT) (err error) {
	dtbBuf := new(bytes.Buffer)
	_, err = fdt.Write(dtbBuf)

	if err != nil {
		return
	}

	image.DeviceTreeBlob = dtbBuf.Bytes()

	return
}

func (image *LinuxImage) fixupBootArgs() (err error) {
	fdt, err := image.fdt()

	if err != nil {
		return
	}

	for _, node := range fdt.RootNode.Children {
		if node.Name == "chosen" {
			bootargs := dt.Property{
				Name:  "bootargs",
				Value: []byte(image.CmdLine + "\x00"),
			}

			node.Properties = append(node.Properties, bootargs)
		}
	}

	return image.updateDTB(fdt)
}

func (image *LinuxImage) fixupInitrd(addr uint32) (err error) {
	fdt, err := image.fdt()

	if err != nil {
		return
	}

	start := addr + uint32(image.InitialRamDiskOffset)
	end := start + uint32(len(image.InitialRamDisk))

	for _, node := range fdt.RootNode.Children {
		if node.Name == "chosen" {
			initrdStart := dt.Property{
				Name:  "linux,initrd-start",
				Value: make([]byte, 4),
			}

			initrdEnd := dt.Property{
				Name:  "linux,initrd-end",
				Value: make([]byte, 4),
			}

			binary.BigEndian.PutUint32(initrdStart.Value, start)
			binary.BigEndian.PutUint32(initrdEnd.Value, end)

			node.Properties = append(node.Properties, initrdStart)
			node.Properties = append(node.Properties, initrdEnd)
		}
	}

	return image.updateDTB(fdt)
}

// Load loads a Linux kernel image in memory.
func (image *LinuxImage) Load() (err error) {
	if image.Region == nil {
		return errors.New("image memory Region must be assigned")
	}

	if len(image.CmdLine) > 0 {
		if len(image.DeviceTreeBlob) == 0 {
			return errors.New("cmdline requires dtb")
		}

		if err = image.fixupBootArgs(); err != nil {
			return fmt.Errorf("cmdline dtb fixup error, %v", err)
		}
	}

	if len(image.InitialRamDisk) > 0 {
		if len(image.DeviceTreeBlob) == 0 {
			return errors.New("initrd requires dtb")
		}

		if err = image.fixupInitrd(image.Region.Start); err != nil {
			return fmt.Errorf("initrd dtb fixup error, %v", err)
		}

		image.Region.Write(image.Region.Start, image.InitialRamDiskOffset, image.InitialRamDisk)
	}

	image.Region.Write(image.Region.Start, image.KernelOffset, image.Kernel)
	image.Region.Write(image.Region.Start, image.DeviceTreeBlobOffset, image.DeviceTreeBlob)

	image.entry = image.Region.Start + uint32(image.KernelOffset)
	image.dtb = image.Region.Start + uint32(image.DeviceTreeBlobOffset)
	image.loaded = true

	return
}

// Entry returns the image entry point.
func (image *LinuxImage) Entry() uint32 {
	return image.entry
}

// Boot calls a loaded Linux kernel image.
func (image *LinuxImage) Boot(cleanup func()) (err error) {
	if !image.loaded {
		return errors.New("Load() kernel before Boot()")
	}

	return boot(image.entry, image.dtb, cleanup)
}
