// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"

	"github.com/u-root/u-root/pkg/boot/bzimage"
)

const (
	// https://docs.kernel.org/arch/x86/boot.html
	minProtocolVersion = 0x0205
)

// LinuxImage represents a bootable Linux kernel image.
type LinuxImage struct {
	// Region is the memory area for image loading.
	Region *dma.Region
	// Memory is the system memory map
	Memory []bzimage.E820Entry

	// Kernel is the Linux kernel image.
	Kernel []byte
	// KernelOffset is the Linux kernel offset from RAM start address.
	KernelOffset int

	// BzImage is the kernel extracted by by Parse() or Load().
	BzImage *bzimage.BzImage

	// InitialRamDisk is the Linux kernel initrd file.
	InitialRamDisk []byte
	// InitialRamDiskOffset is the initrd offset from RAM start address.
	InitialRamDiskOffset int

	// CmdLine is the Linux kernel command line arguments.
	CmdLine string
	// CmdLineOffset is the command line offset from RAM start address.
	CmdLineOffset int

	// ParamsOffset is the boot parameters offset from RAM start address.
	ParamsOffset int

	// DMA pointers
	entry  uint
	params uint
}

// https://docs.kernel.org/arch/x86/zero-page.html
func (image *LinuxImage) buildBootParams() (err error) {
	var buf []byte

	start := image.Region.Start()

	params := &bzimage.LinuxParams{
		MountRootReadonly: 0x01,
		LoaderType:        0xff,
	}

	if n := len(image.CmdLine); n > 0 {
		params.CLPtr = uint32(start) + uint32(image.CmdLineOffset)
		params.CmdLineSize = uint32(n)

		if params.CmdLineSize > image.BzImage.Header.CmdLineSize {
			return errors.New("incompatible command line length")
		}

		image.Region.Write(start, image.CmdLineOffset, []byte(image.CmdLine))
	}

	if len(image.Memory) > bzimage.E820Max {
		return errors.New("image Memory is invalid")
	}

	for i, entry := range image.Memory {
		params.E820Map[i] = entry
		params.E820MapNr += 1
	}

	if n := len(image.InitialRamDisk); n > 0 {
		params.Initrdstart = uint32(start) + uint32(image.InitialRamDiskOffset)
		params.Initrdsize = uint32(n)

		if params.Initrdstart > image.BzImage.Header.InitrdAddrMax {
			return errors.New("incompatible initrd addres")
		}

		image.Region.Write(start, image.InitialRamDiskOffset, image.InitialRamDisk)
	}

	if buf, err = params.MarshalBinary(); err != nil {
		return
	}

	image.Region.Write(start, image.ParamsOffset, buf)

	return
}

func (image *LinuxImage) Parse() (err error) {
	if image.BzImage != nil {
		return
	}

	image.BzImage = &bzimage.BzImage{}

	return image.BzImage.UnmarshalBinary(image.Kernel)
}

// Load loads a Linux kernel image in memory.
func (image *LinuxImage) Load() (err error) {
	if err = image.Parse(); err != nil {
		return
	}

	bzImage := image.BzImage

	if image.Region == nil {
		return errors.New("image memory Region must be assigned")
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

	for _, p := range kelf.Progs {
		buf := make([]byte, p.Filesz)
		p.ReadAt(buf, 0)

		image.Region.Write(image.Region.Start(), image.KernelOffset+int(p.Paddr), buf)
	}

	if err = image.buildBootParams(); err != nil {
		return
	}

	image.entry = image.Region.Start() + uint(image.KernelOffset) + uint(kelf.Entry)
	image.params = image.Region.Start() + uint(image.ParamsOffset)

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
