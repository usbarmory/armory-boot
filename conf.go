// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"log"
)

const defaultConfigPath = "/boot/armory-boot.conf"

var conf Config

type Config struct {
	Debug          bool     `json:"debug"`
	Kernel         []string `json:"kernel"`
	DeviceTreeBlob []string `json:"dtb"`
	CmdLine        string   `json:"cmdline"`
}

//func GetConfig() *Config {
//	return &conf
//}

func (c *Config) Read(partition *Partition, configPath string) (err error) {
	b, err := partition.ReadAll(configPath)

	if err != nil {
		return
	}

	err = json.Unmarshal(b, &c)

	if err != nil {
		return
	}

	if len(conf.Kernel) != 2 {
		return errors.New("invalid kernel parameter size")
	}

	if len(conf.DeviceTreeBlob) != 2 {
		return errors.New("invalid kernel parameter size")
	}

	return
}

func (c *Config) Print() {
	j, _ := json.MarshalIndent(c, "", "\t")

	log.Println("applied configuration:")
	log.Printf("\n%s", string(j))
}
