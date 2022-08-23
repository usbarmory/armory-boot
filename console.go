// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build console
// +build console

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	usbarmory "github.com/usbarmory/tamago/board/usbarmory/mk2"
	"github.com/usbarmory/tamago/soc/nxp/imx6ul"
)

func init() {
	debugConsole, _ := usbarmory.DetectDebugAccessory(250 * time.Millisecond)
	<-debugConsole

	banner := fmt.Sprintf("armory-boot • %s/%s (%s) • %s %s • %s",
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		Revision, Build,
		imx6ul.Model())

	log.SetFlags(0)
	log.Printf("%s", banner)
}
