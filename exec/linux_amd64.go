// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"debug/elf"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"

	"github.com/u-root/u-root/pkg/boot/bzimage"
)

const (
	minProtocolVersion = 0x0205
	cmdLineOffset      = 0
	paramsOffset       = 0x1000
)

// LinuxImage represents a bootable Linux kernel image.
type LinuxImage struct {
	// Memory is the system memory map
	Memory []bzimage.E820Entry
	// Region is the memory area for image loading.
	Region *dma.Region

	// Kernel is the Linux kernel image.
	Kernel []byte

	// CmdLine is the Linux kernel command line arguments.
	CmdLine string

	// DMA pointers
	entry  uint
	params uint

	loaded bool
}

// https://docs.kernel.org/arch/x86/zero-page.html
func (image *LinuxImage) buildBootParams() (addr uint, err error) {
	var buf []byte

	start := image.Region.Start()
	addr = start + paramsOffset

	params := &bzimage.LinuxParams{
		MountRootReadonly: 0x01,
	}

	if n := len(image.CmdLine); n > 0 {
		image.Region.Write(start, cmdLineOffset, []byte(image.CmdLine))

		params.CLPtr = uint32(start) + cmdLineOffset
		params.CmdLineSize = uint32(n)
	}

	for i, entry := range image.Memory {
		params.E820Map[i] = entry
		params.E820MapNr += uint8(i)
	}

	if buf, err = params.MarshalBinary(); err != nil {
		return
	}

	image.Region.Write(start, paramsOffset, buf)

	return
}

// Load loads a Linux kernel image in memory.
func (image *LinuxImage) Load() (err error) {
	if image.Region == nil {
		return errors.New("image memory Region must be assigned")
	}

	bzImage := &bzimage.BzImage{}

	if err = bzImage.UnmarshalBinary(image.Kernel); err != nil {
		return
	}

	if bzImage.Header.Protocolversion < minProtocolVersion {
		return fmt.Errorf("unsupported boot protocol (%v)", bzImage.Header.Protocolversion)
	}

	if bzImage.Header.RelocatableKernel == 0 {
		return errors.New("kernel must be relocatable")
	}

	kelf, err := bzImage.ELF()

	if err != nil {
		return
	}

	image.entry = image.Region.Start() + uint(kelf.Entry)
	start := image.Region.Start()

	for _, section := range kelf.Sections {
		if section.Type != elf.SHT_PROGBITS {
			continue
		}

		offset := uint32(section.Addr)
		offset -= uint32(start)

		i := section.Offset
		j := i + section.Size

		image.Region.Write(start, int(offset), bzImage.KernelCode[i:j])
	}

	if image.params, err = image.buildBootParams(); err != nil {
		return
	}

	image.loaded = true

	return
}

// Entry returns the image entry address.
func (image *LinuxImage) Entry() uint {
	return image.entry
}

// Boot calls a loaded Linux kernel image.
func (image *LinuxImage) Boot(cleanup func()) (err error) {
	if !image.loaded {
		return errors.New("Load() kernel before Boot()")
	}

	return boot(image.entry, image.params, cleanup, nil)
}
