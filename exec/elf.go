// https://github.com/f-secure-foundry/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package exec

import (
	"bytes"
	"debug/elf"
	"fmt"

	"github.com/f-secure-foundry/tamago/dma"
)

// LoadELF implements a _very_ simple ELF loader which is suitable for loading
// bare-metal ELF files like those produced by TamaGo.
func LoadELF(mem *dma.Region, addr uint32, buf []byte) (entry uint32) {
	f, err := elf.NewFile(bytes.NewReader(buf))

	if err != nil {
		panic(err)
	}

	for idx, prg := range f.Progs {
		if prg.Type != elf.PT_LOAD {
			continue
		}

		b := make([]byte, prg.Memsz)

		_, err := prg.ReadAt(b[0:prg.Filesz], 0)

		if err != nil {
			panic(fmt.Sprintf("failed to read LOAD section at idx %d, %q", idx, err))
		}

		offset := uint32(prg.Paddr) - addr
		mem.Write(addr, int(offset), b)
	}

	return uint32(f.Entry)
}

// BootELF loads and boots a unikernel image.
func BootELF(mem *dma.Region, buf []byte, cleanup func()) (err error) {
	entry := LoadELF(mem, mem.Start, buf)
	return boot(entry, 0, cleanup)
}
