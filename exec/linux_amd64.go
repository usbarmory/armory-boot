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

	setupAreaSize = 0x1000000
	cmdLineOffset = 0
	paramsOffset  = 0x1000
)

// LinuxImage represents a bootable Linux kernel image.
type LinuxImage struct {
	// Kernel is the Linux kernel image.
	Kernel []byte

	// Memory is the system memory map
	Memory []bzimage.E820Entry
	// Region is the memory area for image loading.
	Region *dma.Region

	bzImage *bzimage.BzImage
	elf     *elf.File

	// CmdLine is the Linux kernel command line arguments.
	CmdLine string

	// DMA pointers
	entry  uint
	params uint
}

// https://docs.kernel.org/arch/x86/zero-page.html
func (image *LinuxImage) buildBootParams() (addr uint, err error) {
	var buf []byte

	start := image.Region.Start()

	params := &bzimage.LinuxParams{
		MountRootReadonly: 0x01,
	}

	if n := len(image.CmdLine); n > 0 {
		image.Region.Write(start, cmdLineOffset, []byte(image.CmdLine))

		params.CLPtr = uint32(start) + cmdLineOffset
		params.CmdLineSize = uint32(n)
	}

	if len(image.Memory) > bzimage.E820Max {
		return 0, errors.New("image Memory is invalid")
	}

	for i, entry := range image.Memory {
		params.E820Map[i] = entry
		params.E820MapNr += 1
	}

	if buf, err = params.MarshalBinary(); err != nil {
		return 0, err
	}

	addr = start + paramsOffset
	image.Region.Write(start, paramsOffset, buf)

	return
}

// Parse parses a Linux kernel image and returns a suitable memory range for
// its loading Region allocation.
func (image *LinuxImage) Parse() (start uint64, end uint64, err error) {
	bzImage := &bzimage.BzImage{}

	if err = bzImage.UnmarshalBinary(image.Kernel); err != nil {
		return
	}

	if bzImage.Header.Protocolversion < minProtocolVersion {
		return 0, 0, fmt.Errorf("unsupported boot protocol (%v)", bzImage.Header.Protocolversion)
	}

	if bzImage.Header.RelocatableKernel == 0 {
		return 0, 0, errors.New("kernel must be relocatable")
	}

	if image.elf, err = bzImage.ELF(); err != nil {
		return
	}

	start = 0xffffffff

	for _, section := range image.elf.Sections {
		if section.Type != elf.SHT_PROGBITS || section.Size == 0 {
			continue
		}

		// force address space below 4G
		addr := section.Addr & 0xffffffff

		if addr < start {
			start = addr
		}

		if e := addr + section.Size; e > end {
			end = e
		}
	}

	image.entry = uint(start)
	image.bzImage = bzImage

	start -= setupAreaSize

	return
}

// Load loads a Linux kernel in the memory area defined in the Region field.
func (image *LinuxImage) Load() (err error) {
	if image.bzImage == nil {
		if _, _, err = image.Parse(); err != nil {
			return
		}
	}

	if image.Region == nil {
		return errors.New("image memory Region must be assigned")
	}

	start := image.Region.Start()

	for _, section := range image.elf.Sections {
		if section.Type != elf.SHT_PROGBITS || section.Size == 0 {
			continue
		}

		offset := uint32(section.Addr)
		offset -= uint32(start)

		i := section.Offset
		j := i + section.Size

		image.Region.Write(start, int(offset), image.bzImage.KernelCode[i:j])
	}

	if image.params, err = image.buildBootParams(); err != nil {
		return
	}

	return
}

// Entry returns the image entry address.
func (image *LinuxImage) Entry() uint {
	return image.entry
}

// Boot calls a loaded Linux kernel image.
func (image *LinuxImage) Boot(cleanup func()) (err error) {
	if image.params == 0 {
		return errors.New("Load() kernel before Boot()")
	}

	return boot(image.entry, image.params, cleanup, nil)
}
