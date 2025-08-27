Introduction
============

This [TamaGo](https://github.com/usbarmory/tamago) based unikernel
acts as a primary boot loader for the [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki),
allowing boot of kernel images (e.g. Linux) from either the eMMC card or an
external microSD card.

This repository also provides a [command line utility](https://github.com/usbarmory/armory-boot#serial-download-protocol-utility)
to load imx executables through USB using [SDP](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp).

Compiling
=========

Build the [TamaGo compiler](https://github.com/usbarmory/tamago-go)
(or use the [latest binary release](https://github.com/usbarmory/tamago-go/releases/latest)):

```
wget https://github.com/usbarmory/tamago-go/archive/refs/tags/latest.zip
unzip latest.zip
cd tamago-go-latest/src && ./all.bash
cd ../bin && export TAMAGO=`pwd`/go
```

The `BOOT` environment variable must be set to either `uSD` or `eMMC` to
configure the bootloader media for `/boot/armory-boot.conf`, as well as kernel
images, location.

The `START` environment variable must be set to the offset of the first valid
ext4 partition where `/boot/armory-boot.conf` is located (typically 5242880 for
USB armory Mk II default pre-compiled images).

The `CONSOLE` environment variable may be set to `on` to enable serial
logging when a [debug accessory](https://github.com/usbarmory/usbarmory/tree/master/hardware/mark-two-debug-accessory)
is connected.

Build the `armory-boot.imx` application executable:

```
git clone https://github.com/usbarmory/armory-boot && cd armory-boot
make imx BOOT=uSD START=5242880
```

Docker
------

For convenience a docker configuration is provided.

Ensure docker daemon is running and build the `armory-boot` docker image,
this needs to be done only the first time:

```
docker build --build-arg UID=$UID --build-arg GID=$UID -t armory-boot docker
```

You can enter to the container as follows:

```
docker run -it --rm -v $PWD:/build armory-boot
```

Installing
==========

The `armory-boot.imx` file can be flashed on the internal eMMC card or an
external micro SD card as shown in [these instructions](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)#flashing-imx-native-images).

Configuration
=============

The bootloader expects a single configuration file to read information on the
image and parameters to boot.

The bootloader is configured via a single configuration file, and can boot either
 an ARM kernel image or an ELF unikernel (e.g.
[tamago-example](https://github.com/usbarmory/tamago-example)).
The required elements in the configuration file differ depending on the type of
image being loaded, examples for both are given below.

It is an error specify both unikernel and kernel config parameters in the same
configuration file.

Linux kernel boot
-----------------

To load a Linux kernel, the bootloader requires that you provide the paths to
the kernel image and the Device Tree Blob file, along with their respective
SHA256 hashes (only used with configuration signature verification, see _Secure
Boot_), as well as the kernel command line.

An optional initial ramdisk can be passed with the `initrd` parameter.

Example `/boot/armory-boot.conf` configuration file for loading a Linux kernel:

```
{
  "kernel": [
    "/boot/zImage-5.4.51-0-usbarmory",
    "aceb3514d5ba6ac591a7d5f2cad680e83a9f848d19763563da8024f003e927c7"
  ],
  "dtb": [
    "/boot/imx6ulz-usbarmory-default-5.4.51-0.dtb",
    "60d4fe465ef60042293f5723bf4a001d8e75f26e517af2b55e6efaef9c0db1f6"
  ],
  "initrd": [
    "/boot/initrd.img-5.4.51-0-usbarmory",
    "64119096fd329e89f062cb5e0fc5b8e66f98081aef987e0bc7a92a05f4452540"
  ],
  "cmdline": "console=ttymxc1,115200 root=/dev/mmcblk0p1 rootwait rw"
}
```

TamaGo unikernel boot
---------------------

To load a TamaGo unikernel, the bootloader only needs the path to the ELF
binary along with its SHA256 hash (only used with configuration signature
verification, see _Secure Boot_).

Example `/boot/armory-boot.conf` configuration file for loading a TamaGo
unikernel:

```
{
  "unikernel": [
    "/boot/tamago-example",
    "e6de9214249dd7989b4056372424e84b273ff4e5d2410fa12ac230ddaf22690a"
  ]
}
```

Secure Boot
===========

On secure booted systems the `imx_signed` target should be used instead with the relevant
[`HAB_KEYS`](https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)) set.

Additionally, to maintain the chain of trust, the `PUBLIC_KEY` environment
variable must be set with either a [signify](https://man.openbsd.org/signify)
or [minisign](https://jedisct1.github.io/minisign/) public key to enable
configuration file signature verification.

Example key generation (signify):

```
signify -G -p armory-boot.pub -s armory-boot.sec
```

Example key generation (minisign):

```
minisign -G -p armory-boot.pub -s armory-boot.sec
```

Compilation with embedded key:

```
make imx_signed BOOT=uSD START=5242880 PUBLIC_KEY=<last line of armory-boot.pub> HAB_KEYS=<path>
```

When `armory-boot` is compiled with the `PUBLIC_KEY` variable, a signature for
the configuration file must be created in `/boot/armory-boot.conf.sig` using
with the corresponding secret key.

Example signature generation (signify):

```
signify -S -s armory-boot.sec -m armory-boot.conf -x armory-boot.conf.sig
```

Example signature generation (minisign):

```
minisign -S -s armory-boot.sec -m armory-boot.conf -x armory-boot.conf.sig
```

LED status
==========

The [USB armory Mk II](https://github.com/usbarmory/usbarmory/wiki) LEDs
are used, in sequence, as follows:

| Boot sequence                   | Blue | White |
|---------------------------------|------|-------|
| 0. initialization               | off  | off   |
| 1. boot media detected          | on   | off   |
| 2. kernel verification complete | on   | on    |
| 3. jumping to kernel image      | off  | off   |

Serial Download Protocol utility
================================

The `armory-boot-usb` command line utility allows to load an imx executable
through USB using [SDP](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)#serial-download-protocol-sdp),
useful for testing or initial provisioning purposes.

You can automatically download, compile and install the utility, under your
GOPATH, as follows:

```
go install github.com/usbarmory/armory-boot/cmd/armory-boot-usb@latest
```

Alternatively you can manually compile it from source:

```
git clone https://github.com/usbarmory/armory-boot
cd armory-boot && make armory-boot-usb
```

The utility can be cross compiled for Windows as follows:

```
make armory-boot-usb.exe
```

Pre-compiled binaries for Linux and Windows are released
[here](https://github.com/usbarmory/armory-boot/releases).

The utility is meant to be used on devices running in
[USB SDP mode](https://github.com/usbarmory/usbarmory/wiki/Boot-Modes-(Mk-II)):

```
sudo armory-boot-usb -i armory-boot.imx
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
andrea@inversepath.com  

License
=======

armory-boot | https://github.com/usbarmory/armory-boot  
Copyright (c) The armory-boot authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/armory-boot/blob/master/LICENSE) file.
