// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func exec(kernel uint, params uint)
TEXT ·exec(SB),$0-16
	// Disable interrupts
	CLI

	// When booting Linux:
	//   - SI must hold the base address of struct boot_params
	MOVQ	params+8(FP), SI

	// Jump to kernel image
	MOVQ	kernel+0(FP), AX
	JMP	AX
