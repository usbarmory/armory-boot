// Copyright (c) The armory-boot authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

#define SY $0b1111

// func exec(kernel uint, params uint)
TEXT ·exec(SB),$0-16
	MOVD	kernel+0(FP), R16
	MOVD	params+8(FP), R17

	// Disable MMU
	MRS	SCTLR_EL1, R0
	BIC	$1<<0, R0	// disable MMU
	MSR	R0, SCTLR_EL1
	ISB	SY

	// When booting Linux:
	//   - CPU register 0 must be the parameter list address
	//   - CPU register 1 is reserved
	//   - CPU register 2 is reserved
	//   - CPU register 3 is reserved
	MOVD	R17, R0
	MOVD	$0, R1
	MOVD	$0, R2
	MOVD	$0, R3

	// Jump to kernel image
	JMP	(R16)
