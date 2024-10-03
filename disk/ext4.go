// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package disk provides support for SD/MMC card partition access, only ext4
// filesystems are currently supported.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package disk

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dsoprea/go-ext4"

	"github.com/usbarmory/tamago/soc/nxp/usdhc"
)

// Partition represents an SD/MMC card partition, only ext4 filesystems are
// currently supported.
type Partition struct {
	Card    *usdhc.USDHC
	Offset  int64
	_offset int64
}

func (part *Partition) getBlockGroupDescriptor(inode int) (bgd *ext4.BlockGroupDescriptor, err error) {
	_, err = part.Seek(ext4.Superblock0Offset, io.SeekStart)

	if err != nil {
		return
	}

	sb, err := ext4.NewSuperblockWithReader(part)

	if err != nil {
		return
	}

	bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(part, sb)

	if err != nil {
		return
	}

	return bgdl.GetWithAbsoluteInode(inode)
}

func (part *Partition) Read(p []byte) (n int, err error) {
	buf, err := part.Card.Read(part._offset, int64(len(p)))

	if err != nil {
		return
	}

	n = copy(p, buf)
	_, err = part.Seek(int64(n), io.SeekCurrent)

	return
}

func (part *Partition) Seek(offset int64, whence int) (int64, error) {
	info := part.Card.Info()
	end := int64(info.Blocks) * int64(info.BlockSize)

	switch whence {
	case io.SeekStart:
		part._offset = part.Offset + offset
	case io.SeekCurrent:
		part._offset += offset
	case io.SeekEnd:
		part._offset = end + part.Offset + offset
	default:
		return 0, fmt.Errorf("invalid whence %d", whence)
	}

	if part._offset > end {
		return 0, fmt.Errorf("invalid offset %d (%d)", part._offset, offset)
	}

	if part._offset < part.Offset {
		return 0, fmt.Errorf("invalid offset %d (%d)", part._offset, offset)
	}

	return part._offset, nil
}

func (part *Partition) ReadAll(fullPath string) (buf []byte, err error) {
	fullPath = strings.TrimPrefix(fullPath, "/")
	path := strings.Split(fullPath, "/")

	bgd, err := part.getBlockGroupDescriptor(ext4.InodeRootDirectory)

	if err != nil {
		return
	}

	dw, err := ext4.NewDirectoryWalk(part, bgd, ext4.InodeRootDirectory)

	var i int
	var inodeNumber int

	for {
		if err != nil {
			return
		}

		p, de, err := dw.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		deInode := int(de.Data().Inode)

		bgd, err = part.getBlockGroupDescriptor(deInode)

		if err != nil {
			return nil, err
		}

		if p == path[i] {
			if i == len(path)-1 {
				inodeNumber = deInode
				break
			} else {
				dw, err = ext4.NewDirectoryWalk(part, bgd, deInode)

				if err != nil {
					return nil, err
				}

				i += 1
			}
		}
	}

	if inodeNumber == 0 {
		return nil, errors.New("file not found")
	}

	inode, err := ext4.NewInodeWithReadSeeker(bgd, part, inodeNumber)

	if err != nil {
		return
	}

	en := ext4.NewExtentNavigatorWithReadSeeker(part, inode)
	r := ext4.NewInodeReader(en)

	return io.ReadAll(r)
}
