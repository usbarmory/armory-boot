Introduction
============

This [TamaGo](https://github.com/f-secure-foundry/tamago) based unikernel
acts as a primary boot loader for the [USB armory Mk II](https://github.com/f-secure-foundry/usbarmory/wiki),
allowing boot of kernel images (e.g. Linux) from either the eMMC card or an
external microSD card.

Compiling
=========

Build the [TamaGo compiler](https://github.com/f-secure-foundry/tamago-go)
(or use the [latest binary release](https://github.com/f-secure-foundry/tamago-go/releases/latest)):

```
git clone https://github.com/f-secure-foundry/tamago-go -b tamago1.15
cd tamago-go/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

The `BOOT` environment variable must be set to either `uSD` or `eMMC` to
configure the bootloader media for `/boot/armory-boot.conf`, as well as kernel
images, location.

The `START` environment variable must be set to the offset of the first valid
ext4 partition where `/boot/armory-boot.conf` is located (typically 5242880 for
USB armory Mk II default pre-compiled images).

Build the `armory-boot.imx` application executable:

```
git clone https://github.com/f-secure-foundry/armory-boot && cd armory-boot
make CROSS_COMPILE=arm-none-eabi- imx BOOT=uSD START=5242880
```

On secure booted systems the `imx_signed` target should be used to create signed
bootloader images. To maintain the chain of trust, the `PUBLIC_KEY` environment
variable can be set with either a [signify](https://man.openbsd.org/signify) or
[minisign](https://jedisct1.github.io/minisign/) public key to enable
configuration file signature verification.

```
git clone https://github.com/f-secure-foundry/armory-boot && cd armory-boot
make CROSS_COMPILE=arm-none-eabi- imx_signed BOOT=uSD START=5242880 PUBLIC_KEY=RWRss1L3Dg0fATqYIxCuCTgCiaoaIC0vYShv5DuwcXn3c1pDfGvpOY5Q
```

Installing
==========

The `armory-boot.imx` file can be flashed on the internal eMMC card or an
external micro SD card as shown in [these instructions](https://github.com/f-secure-foundry/usbarmory/wiki/Boot-Modes-(Mk-II)#flashing-imx-native-images).

Configuration
=============

The bootloader expects a single configuration file to read information on the
command line, kernel and device tree blob paths along with their SHA256
checksum for validation.

Example `/boot/armory-boot.conf` configuration file:

```
{
  "debug": false,
  "kernel": [
    "/boot/zImage-5.4.51-0-usbarmory",
    "aceb3514d5ba6ac591a7d5f2cad680e83a9f848d19763563da8024f003e927c7"
  ],
  "dtb": [
    "/boot/imx6ulz-usbarmory-default-5.4.51-0.dtb",
    "60d4fe465ef60042293f5723bf4a001d8e75f26e517af2b55e6efaef9c0db1f6"
  ],
  "cmdline": "console=ttymxc1,115200 root=/dev/mmcblk1p1 rootwait rw"
}
```

When `armory-boot` is compiled with the `PUBLIC_KEY` variable, the signature
for the configuration file must be created in `/boot/armory-boot.conf.sig`
using either [signify](https://man.openbsd.org/signify) or
[minisign](https://jedisct1.github.io/minisign/) with the corresponding secret
key.

Example signature generation (signify):

```
signify -G -p armory-boot.pub -s armory-boot.sec
signify -S -s armory-boot.sec -m armory-boot.conf -x armory-boot.conf.sig
```

Example signature generation (minisign):

```
minisign -G -p armory-boot.pub -s armory-boot.sec
minisign -S -s armory-boot.sec -m armory-boot.conf -x armory-boot.conf.sig
```

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

License
=======

armory-boot | https://github.com/f-secure-foundry/armory-boot  
Copyright (c) F-Secure Corporation

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU General Public License as published by the Free Software
Foundation under version 3 of the License.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU General Public License for more details.

See accompanying LICENSE file for full details.
