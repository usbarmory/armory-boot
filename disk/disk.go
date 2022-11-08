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

	"github.com/usbarmory/tamago/soc/nxp/usdhc"
)

// DefaultBootDevice is the default boot device
const DefaultBootDevice = "uSD"

// DefaultOffset is the default start offset of the ext4 partition
const DefaultOffset = 5242880

// Detect initializes the USB armory internal flash ("eMMC") or external
// microSD card ("uSD") as boot device, an ext4 partition must be present at
// the passed start offset. An empty value for device or start parameter selects
// its default value.
func Detect(card *usdhc.USDHC, start string) (part *Partition, err error) {
	offset := int64(DefaultOffset)

	if card == nil {
		return nil, errors.New("invalid card")
	}

	if len(start) > 0 {
		if offset, err = strconv.ParseInt(start, 10, 64); err != nil {
			return nil, fmt.Errorf("invalid start offset, %v\n", err)
		}
	}

	part = &Partition{
		Card:   card,
		Offset: offset,
	}

	if err := part.Card.Detect(); err != nil {
		return nil, fmt.Errorf("could not detect card, %v\n", err)
	}

	return
}
