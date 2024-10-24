//go:build arm64

package macho

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho/types"
)

const (
	sizeOfPointer     = 8
	sizeOfInstruction = 4

	sizeOfStub       = sizeOfInstruction * 3
	sizeOfAuthStub   = sizeOfInstruction * 4
	sizeOfLAResolver = sizeOfPointer
)

func signExtend(val, bits int) int {
	if val&(1<<(bits-1)) != 0 {
		val |= ^0 << bits
	}
	return val
}

// https://developer.arm.com/documentation/ddi0602/2024-09/Base-Instructions/ADRP--Form-PC-relative-address-to-4KB-page-
func adrpImmediate(data []byte) int {
	immLow := (int(data[3]) >> 5) & 0x3
	immHigh := (int(data[0]) >> 5) | (int(data[1]) << 3) | (int(data[2]) << 11)
	return signExtend(immLow|(immHigh<<2), 21) << 12
}

// https://developer.arm.com/documentation/ddi0602/2024-09/Base-Instructions/LDR--immediate---Load-register--immediate--?lang=en
// assume imm12 with unsigned offset variant
func ldrImmediate(data []byte) int {
	scale := data[3] >> 6
	imm12 := (int(data[1]) >> 2) | (int(data[2]&0x3f) << 6)
	return signExtend(imm12, 12) << scale
}

func (me *parser) addArchSpecificSection(parent *contracts.MemoryBlock, sect *types.Section) error {
	// TODO: how do we get the correct names?
	switch sect.Name {
	case "__stubs":
		if sect.Size%sizeOfStub != 0 {
			return fmt.Errorf("invalid stubs section size")
		}
		data, err := sect.Data()
		if err != nil {
			return err
		}
		for i := 0; i < int(sect.Size); i += sizeOfStub {
			// adrp x16, SOME_ADDRESS
			// ldr x16, [x16, SOME_OFFSET]
			// braaz x16
			adrp := data[i : i+sizeOfInstruction]
			ldr := data[i+sizeOfInstruction : i+sizeOfInstruction*2]
			imm := adrpImmediate(adrp)
			imm += ldrImmediate(ldr)
			pc := uintptr(sect.Offset) + uintptr(i)
			pc = pc >> 12 << 12
			// FIXME: address is correct but refers to the final memory instead of binary
			addr := pc + uintptr(imm)

			stub := me.addChild(parent, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("%s stub", "???"),
				Address:      uintptr(sect.Offset) + uintptr(i),
				Size:         uint64(sizeOfStub),
				ParentOffset: uint64(i),
			})
			addValue(stub, "Assembly", addr, 0, 0)
			err = parsingutils.AddLinkWithAddr(stub, "Assembly", "points to", addr)
			if err != nil {
				return err
			}
		}
	case "__auth_stubs":
		if sect.Size%sizeOfAuthStub != 0 {
			return fmt.Errorf("invalid auth stubs section size")
		}
		// TODO: finish?
		// data, err := sect.Data()
		// if err != nil {
		// 	return err
		// }
		for i := 0; i < int(sect.Size); i += sizeOfAuthStub {
			// adrp x17, dyld_internal_addr
			// add x17, x17, dyld_internal_offset
			// ldr x16, [x17]
			// braa x16
			// fmt.Printf("%d: %#x\n", i, data[i:i+sizeOfInstruction])
			// fmt.Printf("%d: %#x\n", i, data[i+sizeOfInstruction:i+sizeOfInstruction*2])
			// fmt.Printf("%d: %#x\n", i, data[i+sizeOfInstruction*2:i+sizeOfInstruction*3])
			// fmt.Printf("%d: %#x\n", i, data[i+sizeOfInstruction*3:i+sizeOfInstruction*4])
			me.addChild(parent, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("%s auth stub", "???"),
				Address:      uintptr(sect.Offset) + uintptr(i),
				Size:         uint64(sizeOfAuthStub),
				ParentOffset: uint64(i),
			})
		}
	case "__la_resolver":
		if sect.Size%sizeOfLAResolver != 0 {
			return fmt.Errorf("invalid la resolver section size")
		}
		for i := 0; i < int(sect.Size); i += sizeOfLAResolver {
			// __stubs point here
			me.addChild(parent, &contracts.MemoryBlock{
				Name:         fmt.Sprintf("%s pointer", "???"),
				Address:      uintptr(sect.Offset) + uintptr(i),
				Size:         uint64(sizeOfPointer),
				ParentOffset: uint64(i),
			})
		}
	}
	return nil
}
