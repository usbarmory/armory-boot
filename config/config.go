// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package config provides parsing for the armory-boot configuration file
// format.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/f-secure-foundry/armory-boot/disk"
)

// DefaultConfigPath is the default armory-boot configuration file path.
const DefaultConfigPath = "/boot/armory-boot.conf"

// DefaultSignaturePath is the default armory-boot configuration file signature
// path.
const DefaultSignaturePath = "/boot/armory-boot.conf.sig"

// Config represents the armory-boot configuration.
type Config struct {
	// KernelPath is the path to a Linux kernel image.
	KernelPath []string `json:"kernel"`

	// DeviceTreeBlobPath is the path to a Linux DTB file.
	DeviceTreeBlobPath []string `json:"dtb"`

	// InitialRamDiskPath is the path to a Linux initrd file.
	InitialRamDiskPath []string `json:"initrd"`

	// CmdLine is the Linux kernel command-line parameters.
	CmdLine string `json:"cmdline"`

	// Unikernel is the path to an ELF unikernel image (e.g. TamaGo).
	UnikernelPath []string `json:"unikernel"`

	// ELF indicates whether the loaded kernel is a unikernel or not.
	ELF bool

	// JSON holds the configuration file contents
	JSON []byte

	kernel []byte
	dtb    []byte
	initrd []byte

	kernelHash string
	dtbHash    string
	initrdHash string
}

func (c *Config) init(part *disk.Partition) (err error) {
	var kernelPath string

	if err = json.Unmarshal(c.JSON, &c); err != nil {
		return
	}

	ul, kl := len(c.UnikernelPath), len(c.KernelPath)
	isUnikernel, isKernel := ul > 0, kl > 0

	if isUnikernel == isKernel {
		return errors.New("must specify either unikernel or kernel")
	}

	switch {
	case isKernel:
		if kl != 2 {
			return errors.New("invalid kernel parameter size")
		}

		if len(c.DeviceTreeBlobPath) != 2 {
			return errors.New("invalid dtb parameter size")
		}

		if len(c.InitialRamDiskPath) > 0 {
			if len(c.InitialRamDiskPath) != 2 {
				return errors.New("invalid initrd parameter size")
			}

			if c.initrd, err = part.ReadAll(c.InitialRamDiskPath[0]); err != nil {
				return
			}

			c.initrdHash = c.InitialRamDiskPath[1]
		}

		kernelPath = c.KernelPath[0]
		c.kernelHash = c.KernelPath[1]

		if c.dtb, err = part.ReadAll(c.DeviceTreeBlobPath[0]); err != nil {
			return
		}

		c.dtbHash = c.DeviceTreeBlobPath[1]
	case isUnikernel:
		if ul != 2 {
			return errors.New("invalid unikernel parameter size")
		}

		kernelPath = c.UnikernelPath[0]
		c.kernelHash = c.UnikernelPath[1]
	}

	if c.kernel, err = part.ReadAll(kernelPath); err != nil {
		return fmt.Errorf("invalid path %s, %v", kernelPath, err)
	}

	if err != nil {
		return fmt.Errorf("invalid path %s, %v", c.DeviceTreeBlobPath[0], err)
	}

	if isUnikernel {
		c.ELF = true
	}

	return
}

// Load reads an armory-boot configuration file, and optionally its signature,
// from a disk partition. The public key argument is used for signature
// authentication, a valid signature path must be present if a key is set.
func Load(part *disk.Partition, configPath string, sigPath string, pubKey string) (c *Config, err error) {
	log.Printf("armory-boot: loading configuration at %s\n", configPath)

	c = &Config{}

	if c.JSON, err = part.ReadAll(configPath); err != nil {
		return
	}

	if len(pubKey) > 0 {
		sig, err := part.ReadAll(sigPath)

		if err != nil {
			return nil, fmt.Errorf("invalid signature path, %v", err)
		}

		if err = Verify(c.JSON, sig, pubKey); err != nil {
			return nil, err
		}
	}

	defer func() {
		if err != nil {
			c.kernel = nil
			c.dtb = nil
			c.initrd = nil
		}
	}()

	if err = c.init(part); err != nil {
		return
	}

	if len(pubKey) == 0 {
		return
	}

	if !CompareHash(c.kernel, c.kernelHash) {
		err = errors.New("invalid kernel hash")
		return
	}

	if len(c.dtb) > 0 && !CompareHash(c.dtb, c.dtbHash) {
		err = errors.New("invalid dtb hash")
		return
	}

	if len(c.initrd) > 0 && !CompareHash(c.initrd, c.initrdHash) {
		err = errors.New("invalid initrd hash")
		return
	}

	return
}

// Kernel returns the contents of the kernel image previously loaded by a
// successful Load().
func (c *Config) Kernel() []byte {
	return c.kernel
}

// DeviceTreeBlob returns the contents of the dtb file previously loaded by a
// successful Load().
func (c *Config) DeviceTreeBlob() []byte {
	return c.dtb
}

// InitialRamDisk returns the contents of the initrd image previously loaded by
// a successful Load().
func (c *Config) InitialRamDisk() []byte {
	return c.initrd
}
