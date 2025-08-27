// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package sdp provide helpers for implementing the Serial Download Protocol
// (SDP), used on NXP i.MX System-on-Chip (SoC) application processors.
package sdp

import (
	"bytes"
	"encoding/binary"
)

const HIDReportSize = 1024

// SDP command types
// (p322, 8.9.3 Serial Download Protocol (SDP), IMX6ULLRM).
const (
	ReadRegister  = 0x0101
	WriteFile     = 0x0404
	DCDWrite      = 0x0a0a
	JumpAddress   = 0x0b0b
	SkipDCDHeader = 0x0c0c
)

// SDP represents an SDP command
// (p322, 8.9.3 Serial Download Protocol (SDP), IMX6ULLRM).
type SDP struct {
	CommandType uint16
	Address     uint32
	Format      uint8
	DataCount   uint32
	Data        uint32
	_           byte
}

// Bytes converts the SDP command structure to byte array format.
func (s *SDP) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, s)
	return buf.Bytes()
}

// BuildReadRegisterReport generates USB HID reports (ID 1) that implement the
// SDP READ_REGISTER command for reading a single 32-bit device register value
// (p323, 8.9.3.1.1 READ_REGISTER, IMX6ULLRM).
func BuildReadRegisterReport(addr uint32, size uint32) (r1 []byte) {
	sdp := &SDP{
		CommandType: ReadRegister,
		Address:     addr,
		Format:      0x20, // 32-bit access
		DataCount:   size,
	}

	return sdp.Bytes()
}

// BuildDCDWriteReport generates USB HID reports (IDs 1 and 2) that implement
// the SDP DCD_WRITE command sequence, used to load a DCD binary payload
// (p327, 8.9.3.1.5 DCD_WRITE, IMX6ULLRM).
func BuildDCDWriteReport(dcd []byte, addr uint32) (r1 []byte, r2 []byte) {
	sdp := &SDP{
		CommandType: DCDWrite,
		Address:     addr,
		DataCount:   uint32(len(dcd)),
	}

	return sdp.Bytes(), dcd
}

// BuildFileWriteReport generates USB HID reports (IDs1 and 2) that implement
// the SDP FILE_WRITE command sequence, used to load an imx binary payload
// (p325, 8.9.3.1.3 FILE_WRITE, IMX6ULLRM).
func BuildFileWriteReport(imx []byte, addr uint32) (r1 []byte, r2 [][]byte) {
	sdp := &SDP{
		CommandType: WriteFile,
		Address:     addr,
		DataCount:   uint32(len(imx)),
	}

	// make a copy to leave input slice untouched
	imx = append(imx[:0:0], imx...)

	// DCD pointer must be cleared
	binary.LittleEndian.PutUint32(imx[DCDOffset:], 0)

	for j := 0; j < len(imx); j += HIDReportSize {
		k := j + HIDReportSize

		if k > len(imx) {
			k = len(imx)
		}

		r2 = append(r2, imx[j:k])
	}

	return sdp.Bytes(), r2
}

// BuildJumpAddressReport generates the USB HID report (ID 1) that implements
// the SDP JUMP_ADDRESS command, used to execute an imx binary payload
// (p328, 8.9.3.1.7 JUMP_ADDRESS, IMX6ULLRM).
func BuildJumpAddressReport(addr uint32) (r1 []byte) {
	sdp := &SDP{
		CommandType: JumpAddress,
		Address:     addr,
	}

	return sdp.Bytes()
}
