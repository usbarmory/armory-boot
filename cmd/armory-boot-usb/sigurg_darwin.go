// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"os/signal"
	"syscall"
)

func init() {
	// Workaround for macOS HID API calls (IOHIDDeviceSetReport) failing
	// with a general error when interrupted by a signal. SIGURG is the most
	// frequent offender due to Go's asynchronous preemption mechanism, but
	// the underlying issue is likely that the HID library doesn't handle
	// EINTR or signals in general.
	// See https://github.com/sstallion/go-hid/issues/15.
	signal.Ignore(syscall.SIGURG)
}
