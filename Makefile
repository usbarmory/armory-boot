# Copyright (c) The armory-boot authors. All Rights Reserved.
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart,linkprintk
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

SHELL = /bin/bash
START ?= 5242880

ifeq ("${CONSOLE}","on")
	BUILD_TAGS := ${BUILD_TAGS},console
endif

APP := armory-boot
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOOSPKG=github.com/usbarmory/tamago GOARM=7 GOARCH=arm
TEXT_START := 0x90010000 # ramStart (defined in imx6/imx6ul/memory.go) + 0x10000
TAMAGOFLAGS := -tags ${BUILD_TAGS} -trimpath -ldflags "-T $(TEXT_START) -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}' -X 'main.Boot=${BOOT}' -X 'main.Start=${START}' -X 'main.PublicKeyStr=${PUBLIC_KEY}'"
GOFLAGS := -trimpath -ldflags "-s -w"

.PHONY: clean

#### primary targets ####

all: $(APP)

imx: $(APP).imx

imx_signed: $(APP)-signed.imx

elf: $(APP)

$(APP)-usb:
	@if [ "${TAMAGO}" != "" ]; then \
		${TAMAGO} build $(GOFLAGS) cmd/$(APP)-usb/*.go; \
	else \
		go build $(GOFLAGS) cmd/$(APP)-usb/*.go; \
	fi

$(APP)-usb.exe: BUILD_OPTS := GOOS=windows CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc
$(APP)-usb.exe:
	@if [ "${TAMAGO}" != "" ]; then \
		$(BUILD_OPTS) ${TAMAGO} build cmd/$(APP)-usb/$(APP)-usb.go; \
	else \
		$(BUILD_OPTS) go build cmd/$(APP)-usb/$(APP)$(APP)-usb.go; \
	fi

#### utilities ####

check_env:
	@if [ "${BOOT}" != "eMMC" ] && [ "${BOOT}" != "uSD" ]; then \
		echo 'You need to set the BOOT variable to either eMMC or uSD to select boot media'; \
		exit 1; \
	fi

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

check_hab_keys:
	@if [ "${HAB_KEYS}" == "" ]; then \
		echo 'You need to set the HAB_KEYS variable to the path of secure boot keys'; \
		echo 'See https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)'; \
		exit 1; \
	fi

dcd:
	echo $(GOMODCACHE)
	echo $(TAMAGO_PKG)
	cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/usbarmory/mk2/imximage.cfg $(APP).dcd

clean:
	@rm -fr $(APP) $(APP).bin $(APP).imx $(APP)-signed.imx $(APP).csf $(APP).dcd $(APP)-usb $(APP)-usb.exe

#### dependencies ####

$(APP): check_tamago check_env
	$(GOENV) $(TAMAGO) build $(TAMAGOFLAGS) -o ${APP}

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell ${TAMAGO} list -m -f '{{.Path}}@{{.Version}}' github.com/usbarmory/tamago)
$(APP).dcd: dcd

$(APP).bin: CROSS_COMPILE=arm-none-eabi-
$(APP).bin: $(APP)
	$(CROSS_COMPILE)objcopy --enable-deterministic-archives \
	    -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP) -O binary $(APP).bin

$(APP).imx: SOURCE_DATE_EPOCH=0
$(APP).imx: $(APP).bin $(APP).dcd
	mkimage -n $(APP).dcd -T imximage -e $(TEXT_START) -d $(APP).bin $(APP).imx
	# Copy entry point from ELF file
	dd if=$(APP) of=$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc

#### secure boot ####

$(APP)-signed.imx: check_hab_keys $(APP).imx
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool@latest
	$(shell ${TAMAGO} env GOPATH)/bin/habtool \
		-A ${HAB_KEYS}/CSF_1_key.pem \
		-a ${HAB_KEYS}/CSF_1_crt.pem \
		-B ${HAB_KEYS}/IMG_1_key.pem \
		-b ${HAB_KEYS}/IMG_1_crt.pem \
		-t ${HAB_KEYS}/SRK_1_2_3_4_table.bin \
		-x 1 \
		-i $(APP).imx \
		-o $(APP).csf && \
	cat $(APP).imx $(APP).csf > $(APP)-signed.imx
