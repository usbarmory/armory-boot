// https://github.com/usbarmory/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package sdp

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Program image tags
// (p309, 8.7 Program image, IMX6ULLRM).
const (
	TagIVT = 0xd1
	TagDCD = 0xd2
)

// DCD constants
// (p312, 8.7.2 Device Configuration Data (DCD), IMX6ULLRM).
const (
	// write command tag
	WriteData = 0xcc
	// DCD pointer offset within IVT
	DCDOffset = 12
	// maximum DCD size
	DCDSize = 1768
)

// DCDHeader represents a DCD header
// (p312, 8.7.2 Device Configuration Data (DCD), IMX6ULLRM).
type DCDHeader struct {
	Tag     uint8
	Length  uint16
	Version uint8
}

// IVT represents an IVT entry
// (p311, 8.7.1.1 Image vector table structure, IMX6ULLRM).
type IVT struct {
	Tag      uint8
	Length   uint16
	Version  uint8
	Entry    uint32
	_        uint32
	DCD      uint32
	BootData uint32
	Self     uint32
	CSF      uint32
	_        uint32
}

// ParseIVT extracts the Image Vector Table (IVT) from an imx format binary
// image.
func ParseIVT(imx []byte) (ivt *IVT, err error) {
	ivt = &IVT{}

	if err = binary.Read(bytes.NewReader(imx), binary.LittleEndian, ivt); err != nil {
		return nil, err
	}

	if ivt.Tag != TagIVT {
		return nil, errors.New("could not find IVT tag")
	}

	return
}

// ParseDCD extracts the Device Configuration Data (DCD) from an imx format
// binary image.
func ParseDCD(imx []byte, ivt *IVT) (dcd []byte, err error) {
	hdr := &DCDHeader{}
	dcdStart := ivt.DCD - ivt.Self

	if len(imx) < int(dcdStart+4) {
		return nil, errors.New("could not parse DCD, insufficient length")
	}

	if err = binary.Read(bytes.NewReader(imx[dcdStart:dcdStart+4]), binary.BigEndian, hdr); err != nil {
		return
	}

	if hdr.Tag != TagDCD {
		return nil, errors.New("could not find DCD tag")
	}

	if hdr.Length > DCDSize || int(hdr.Length) > int(dcdStart)+len(imx) {
		return nil, errors.New("could not parse DCD, invalid length")
	}

	dcd = imx[dcdStart : dcdStart+uint32(hdr.Length)]

	return
}
