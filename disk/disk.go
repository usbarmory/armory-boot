// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package disk

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/f-secure-foundry/tamago/board/f-secure/usbarmory/mark-two"
)

// Detect initializes the USB armory internal flash ("eMMC") or external
// microSD card ("uSD") as boot device, an ext4 partition must be present at
// the passed start offset.
func Detect(device string, start string) (part *Partition, err error) {
	offset, err := strconv.ParseInt(start, 10, 64)

	if err != nil {
		return nil, fmt.Errorf("invalid start offset, %v\n", err)
	}

	part = &Partition{
		Offset: offset,
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
