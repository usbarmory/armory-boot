// https://github.com/usbarmory/armory-boot
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// func svc()
TEXT ·svc(SB),$0
	SWI	$0

// func exec(kernel uint32, params uint32)
TEXT ·exec(SB),$0-8
	MOVW	kernel+0(FP), R3
	MOVW	params+4(FP), R4

	// Disable MMU
	MRC	15, 0, R0, C1, C0, 0
	BIC	$1, R0
	MCR	15, 0, R0, C1, C0, 0

	// CPU register 0 must be 0
	MOVW	$0, R0
	// CPU register 1 not required for DTB boot
	// CPU register 2 must be the parameter list address
	MOVW	R4, R2

	// Jump to kernel image
	B	(R3)
