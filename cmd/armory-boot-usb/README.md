Introduction
============

This tool implements a minimal set of the Serial Download Protocol (SDP), used
on NXP i.MX SoC application processors, to load an executable image over USB.

The tool is aimed at [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki) users but should work
with all SoCs from the i.MX series capable of USB HID based SDP (only tested
devices are listed as supported, Pull Requests are welcome to expand this set).

The [mfgtools](https://github.com/NXPmicro/mfgtools) and
[imx_usb_loader](https://github.com/boundarydevices/imx_usb_loader) projects
also implement similar functionality.

Serial Download Protocol
========================

The `armory-boot-usb` command line utility allows to load an imx executable
through USB using [SDP](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp),
useful for testing or initial provisioning purposes.

The utility can be compiled installed in your $GOPATH as follows:

```
go get github.com/f-secure-foundry/armory-boot/cmd/armory-boot-usb
```

Alternatively pre-compiled binaries for Linux and Windows are released
[here](https://github.com/f-secure-foundry/armory-boot/releases).

It can then be used on devices running in
[USB SDP mode](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)):

```
sudo ./armory-boot-usb -i armory-boot.imx
found device 15a2:0080 Freescale SemiConductor Inc  SE Blank 6ULL
parsing armory-boot.imx
loading DCD at 0x00910000 (952 bytes)
loading imx to 0x9000f400 (2182144 bytes)
jumping to 0x9000f400
serial download complete
```

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

License
=======

armory-boot | https://github.com/f-secure-foundry/armory-boot  
Copyright (c) F-Secure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/f-secure-foundry/armory-boot/blob/master/LICENSE) file.
