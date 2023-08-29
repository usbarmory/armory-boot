// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// This tool implements a minimal set of the Serial Download Protocol (SDP),
// used on NXP i.MX SoC application processors, to load an executable image
// over USB.
//
// It implements a subset of the functionality also available in the following
// tools:
//  * https://github.com/NXPmicro/mfgtools
//  * https://github.com/boundarydevices/imx_usb_loader

package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/usbarmory/armory-boot/sdp"

	"github.com/usbarmory/hid"
)

const (
	// USB vendor ID for all supported devices
	FreescaleVendorID = 0x15a2

	// On-Chip RAM (OCRAM/iRAM) address for payload staging
	iramOffset = 0x00910000
)

// SDP HID report IDs
// (p327, 8.9.3.1 SDP commands, IMX6ULLRM).
const (
	H2D_COMMAND       = 1 // Command  - Host to Device
	H2D_DATA          = 2 // Data     - Host to Device
	D2H_RESPONSE      = 3 // Response - Device to Host
	D2H_RESPONSE_LAST = 4 // Response - Device to Host
)

// This tool should work with all SoCs from the i.MX series capable of USB HID
// based SDP, only tested devices are listed as supported, Pull Requests are
// welcome to expand this set.
var supportedDevices = map[uint16]string{
	0x007d: "Freescale SemiConductor Inc  SE Blank 6UL",
	0x0080: "Freescale SemiConductor Inc  SE Blank 6ULL",
}

type Config struct {
	input   string
	timeout int
	dev     hid.Device
}

var conf *Config

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	conf = &Config{}

	flag.IntVar(&conf.timeout, "t", 5, "timeout in seconds for command responses")
	flag.StringVar(&conf.input, "i", "", "imx file")
}

// detect compatible devices in SDP mode
func detect() (err error) {
	devices, err := hid.Devices()

	if err != nil {
		return
	}

	for _, d := range devices {
		if d.VendorID != FreescaleVendorID {
			continue
		}

		if product, ok := supportedDevices[d.ProductID]; ok {
			log.Printf("found device %04x:%04x %s", d.VendorID, d.ProductID, product)
		} else {
			continue
		}

		conf.dev, err = d.Open()

		if err != nil {
			return
		}

		break
	}

	if conf.dev == nil {
		return errors.New("no device found, target missing or invalid permissions (forgot admin shell?)")
	}

	return
}

func sendHIDReport(reqID int, buf []byte, resID int) (res []byte, err error) {
	err = conf.dev.Write(append([]byte{byte(reqID)}, buf...))

	if err != nil || resID < 0 {
		return
	}

	ok := false
	timer := time.After(time.Duration(conf.timeout) * time.Second)

	for {
		select {
		case res, ok = <-conf.dev.ReadCh():
			if !ok {
				return nil, errors.New("error reading response")
			}

			if len(res) > 0 && res[0] == byte(resID) {
				return
			}
		case <-timer:
			return nil, errors.New("command timeout")
		}
	}
}

func dcdWrite(dcd []byte, addr uint32) (err error) {
	r1, r2 := sdp.BuildDCDWriteReport(dcd, addr)

	_, err = sendHIDReport(H2D_COMMAND, r1, -1)

	if err != nil {
		return
	}

	_, err = sendHIDReport(H2D_DATA, r2, D2H_RESPONSE)

	return
}

func fileWrite(imx []byte, addr uint32) (err error) {
	r1, r2 := sdp.BuildFileWriteReport(imx, addr)

	_, err = sendHIDReport(H2D_COMMAND, r1, -1)

	if err != nil {
		return
	}

	resID := -1
	timer := time.After(time.Duration(conf.timeout) * time.Second)

	for i, r := range r2 {
		if i == len(r2)-1 {
			resID = D2H_RESPONSE_LAST
		}
	send:
		_, err = sendHIDReport(H2D_DATA, r, resID)

		if err != nil && runtime.GOOS == "darwin" && err.Error() == "hid: general error" {
			// On macOS access contention with the OS causes
			// errors, as a workaround we retry from the transfer
			// that got caught up.
			select {
			case <-timer:
				return
			default:
				off := uint32(i) * 1024
				r1 := &sdp.SDP{
					CommandType: sdp.WriteFile,
					Address:     addr + off,
					DataCount:   uint32(len(imx)) - off,
				}

				if _, err = sendHIDReport(H2D_COMMAND, r1.Bytes(), -1); err != nil {
					return
				}

				goto send
			}
		}

		if err != nil {
			break
		}
	}

	return
}

func jumpAddress(addr uint32) (err error) {
	r1 := sdp.BuildJumpAddressReport(addr)
	_, err = sendHIDReport(H2D_COMMAND, r1, -1)

	return
}

func main() {
	var err error

	flag.Parse()

	if len(conf.input) <= 0 {
		flag.PrintDefaults()
		return
	}

	if err = detect(); err != nil {
		log.Fatal(err)
	}

	log.Printf("parsing %s", conf.input)
	imx, err := ioutil.ReadFile(conf.input)

	if err != nil {
		log.Fatal(err)
	}

	ivt, err := sdp.ParseIVT(imx)

	if err != nil {
		log.Fatalf("IVT parsing error: %v", err)
	}

	dcd, err := sdp.ParseDCD(imx, ivt)

	if err != nil {
		log.Fatalf("DCD parsing error: %v", err)
	}

	log.Printf("loading DCD at %#08x (%d bytes)", iramOffset, len(dcd))
	err = dcdWrite(dcd, iramOffset)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("loading imx to %#08x (%d bytes)", ivt.Self, len(imx))
	err = fileWrite(imx, ivt.Self)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("jumping to %#08x", ivt.Self)
	err = jumpAddress(ivt.Self)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("serial download complete")
}
