// Copyright (c) WithSecure Corporation
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/usbarmory/tamago/dma"

	"github.com/u-root/u-root/pkg/boot/bzimage"
)

// Linux kernel information
const (
	// https://docs.kernel.org/arch/x86/boot.html
	MinProtocolVersion = 0x0205
	// https://github.com/torvalds/linux/blob/master/include/uapi/linux/screen_info.h
	VideoTypeEFI = 0x70
)

// Zero page offsets (https://docs.kernel.org/arch/x86/zero-page.html)
const (
	screenInfoOffset = 0x00
	efiInfoOffset    = 0x1c0
)

// EFI Information (efi_info) signatures
var (
	EFI64LoaderSignature = [4]byte{0x45, 0x4c, 0x36, 0x34} // "EL64"
	EFI32LoaderSignature = [4]byte{0x45, 0x4c, 0x33, 0x32} // "EL32"
)

// EFI represents the Linux Zero Page `efi_info` structure.
type EFI struct {
	LoaderSignature   [4]byte
	SystemTable       uint32
	MemoryDescSize    uint32
	MemoryDescVersion uint32
	MemoryMap         uint32
	MemoryMapSize     uint32
	SystemTableHigh   uint32
	MemoryMapHigh     uint32
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *EFI) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

// Screen represents the Linux Zero Page `screen_info` structure.
type Screen struct {
	OrigX           uint8
	OrigY           uint8
	ExtMemK         uint16
	OrigVideoPage   uint16
	OrigVideoMode   uint8
	OrigVideoCols   uint8
	_               uint16
	OrigVideoeGabx  uint16
	_               uint16
	OrigVideoLines  uint8
	OrigVideoIsVGA  uint8
	OrigVideoPoints uint16
	LfbWidth        uint16
	LfbHeight       uint16
	LfbDepth        uint16
	LfbBase         uint32
	LfbSize         uint32
	CLMagic         uint16
	CLOffset        uint16
	LfbLineLength   uint16
	RedSize         uint8
	RedPos          uint8
	GreenSize       uint8
	GreenPos        uint8
	BlueSize        uint8
	BluePos         uint8
	RsvdSize        uint8
	RsvdPos         uint8
	VesapmSeg       uint16
	VesapmOff       uint16
	Pages           uint16
	VesaAttributes  uint16
	Capabilities    uint32
	ExtLfbBase      uint32
	_               [2]uint8
}

// MarshalBinary implements the [encoding.BinaryMarshaler] interface.
func (d *Screen) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, d)
	return buf.Bytes(), nil
}

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

	// EFI is the boot parameters EFI information.
	EFI *EFI
	// Screen is the boot parameters frame buffer information.
	Screen *Screen

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

		image.Region.Write(start, image.CmdLineOffset, []byte(image.CmdLine + "\x00"))
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

	if image.EFI != nil {
		if buf, err = image.EFI.MarshalBinary(); err != nil {
			return
		}

		image.Region.Write(start, image.ParamsOffset+efiInfoOffset, buf)
	}

	if image.Screen != nil {
		if buf, err = image.Screen.MarshalBinary(); err != nil {
			return
		}

		image.Region.Write(start, image.ParamsOffset+screenInfoOffset, buf)
	}

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

	if bzImage.Header.Protocolversion < MinProtocolVersion {
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
