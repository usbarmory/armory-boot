// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package disk

import (
	"errors"
	"fmt"
	"strconv"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
)

// DefaultBootDevice is the default boot device
const DefaultBootDevice = "uSD"

// DefaultOffset is the default start offset of the ext4 partition
const DefaultOffset = 5242880

// Detect initializes the USB armory internal flash ("eMMC") or external
// microSD card ("uSD") as boot device, an ext4 partition must be present at
// the passed start offset. An empty value for device or start parameter selects
// its default value.
func Detect(device string, start string) (part *Partition, err error) {
	offset := int64(DefaultOffset)

	if len(start) > 0 {
		if offset, err = strconv.ParseInt(start, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid start offset, %v\n", err)
		}
	}

	part = &Partition{
		Offset: offset,
	}

	if len(device) == 0 {
		device = DefaultBootDevice
	}

	switch device {
	case "eMMC":
		part.Card = usbarmory.MMC
	case "uSD":
		part.Card = usbarmory.SD
	default:
		return nil, errors.New("invalid boot parameter")
	}

	if err := part.Card.Detect(); err != nil {
		return nil, fmt.Errorf("could not detect %s, %v\n", device, err)
	}

	return
}
