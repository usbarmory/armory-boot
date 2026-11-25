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

// func exec_efi(entry uint, handle uint, table uint)
TEXT ·exec_efi(SB),$0-24
	// Disable interrupts
	CLI

	// When booting EFI images the handoff state needs:
	//   - CX must hold EFI_HANDLE
	//   - DX must hold EFI_SYSTEM_TABLE*
	MOVQ	handle+8(FP), CX
	MOVQ	table+16(FP), DX

	// Jump to kernel image
	MOVQ	kernel+0(FP), AX
	JMP	AX
