// https://github.com/usbarmory/armory-boot
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func svc()
TEXT ·svc(SB),$0
	SWI	$0

// func exec(kernel uint32, params uint32, mmu bool)
TEXT ·exec(SB),$0-12
	MOVW	kernel+0(FP), R3
	MOVW	params+4(FP), R4
	MOVW	mmu+8(FP), R5

	CMP	$1, R5
	B.EQ	mmu

	// Disable MMU
	MRC	15, 0, R0, C1, C0, 0
	BIC	$1, R0
	MCR	15, 0, R0, C1, C0, 0
mmu:
	// When booting Linux:
	//   - CPU register 0 must be 0
	//   - CPU register 1 not required for DTB boot
	//   - CPU register 2 must be the parameter list address
	//   - CPU must be in SVC mode
	MOVW	$0, R0
	MOVW	R4, R2

	// switch to supervisor mode
	WORD	$0xf1020013	// cps 0x13

	// Jump to kernel image
	B	(R3)
