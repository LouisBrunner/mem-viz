package parse

import (
	"fmt"
	"io"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"golang.org/x/exp/slices"
)

func isUnslidAddress[A addressOrOffset](special A) bool {
	_, ok := interface{}(special).(subcontracts.UnslidAddress)
	return ok
}

func calculateAddress[A addressOrOffset](base uintptr, special A, slide uint64) uintptr {
	address := base + uintptr(special)
	if isUnslidAddress(special) {
		address = uintptr(uint64(special) + slide)
		fmt.Printf("Unslid address %#16x + %#16x = %#16x\n", special, slide, address)
	}
	return address
}

func addParentOffset[A addressOrOffset](special A, parent uint64) A {
	if isUnslidAddress(special) {
		return special
	}
	return A(uint64(special) + parent)
}

func getReaderAtOffset[A addressOrOffset](cache subcontracts.Cache, special A, offset uint64) io.Reader {
	addr := calculateAddress(0, special, gSlide) + uintptr(offset)
	if isUnslidAddress(special) {
		return cache.ReaderAbsolute(uint64(addr))
	}
	return cache.ReaderAtOffset(int64(addr))
}

func createCommonBlock[A addressOrOffset](parent *contracts.MemoryBlock, label string, offset A, size uint64) (*contracts.MemoryBlock, error) {
	address := calculateAddress(parent.Address, offset, gSlide)
	if address < parent.Address {
		return nil, fmt.Errorf("address of %q (%#16x) is before parent %q (%#16x)", label, address, parent.Name, parent.Address)
	}
	block := &contracts.MemoryBlock{
		Name:         label,
		Address:      address,
		Size:         size,
		ParentOffset: uint64(address - parent.Address),
	}
	addChild(parent, block)
	return block, nil
}

func addChild(parent, child *contracts.MemoryBlock) {
	for i, curr := range parent.Content {
		if curr.Address > child.Address {
			parent.Content = slices.Insert(parent.Content, i, child)
			return
		}
	}
	parent.Content = append(parent.Content, child)
}

func findAndRemoveChild(parent, child *contracts.MemoryBlock) error {
	for i, curr := range parent.Content {
		if curr == child {
			removeChild(parent, i)
			return nil
		}
	}
	return fmt.Errorf("could not find child %+v in parent %+v", child, parent)
}

func removeChild(parent *contracts.MemoryBlock, childIndex int) *contracts.MemoryBlock {
	child := parent.Content[childIndex]
	parent.Content = slices.Delete(parent.Content, childIndex, childIndex+1)
	return child
}

func moveChild(parent, newParent *contracts.MemoryBlock, childIndex int) {
	child := removeChild(parent, childIndex)
	child.ParentOffset = uint64(child.Address - newParent.Address)
	addChild(newParent, child)
}

func addLink(parent *contracts.MemoryBlock, parentValueName string, child *contracts.MemoryBlock, linkName string) error {
	for i := range parent.Values {
		if parent.Values[i].Name != parentValueName {
			continue
		}

		parent.Values[i].Links = append(parent.Values[i].Links, &contracts.MemoryLink{
			Name:          linkName,
			TargetAddress: uint64(child.Address),
		})
		return nil
	}

	return fmt.Errorf("could not find value %q in parent %+v", parentValueName, parent)
}

func addLinkWithOffset[A addressOrOffset](parent *contracts.MemoryBlock, parentValueName string, offset A, linkName string) error {
	if offset == 0 {
		return nil
	}

	for i := range parent.Values {
		if parent.Values[i].Name != parentValueName {
			continue
		}

		parent.Values[i].Links = append(parent.Values[i].Links, &contracts.MemoryLink{
			Name:          linkName,
			TargetAddress: uint64(calculateAddress(parent.Address, offset, gSlide)),
		})
		return nil
	}

	return fmt.Errorf("could not find value %q in parent %+v", parentValueName, parent)
}

func formatValue(name string, value interface{}) string {
	formatInteger := func() string {
		// FIXME: would be nice to have more specialized formats, e.g. for OSVersion or CacheType
		return fmt.Sprintf("%#x", value)
	}

	switch v := value.(type) {
	case subcontracts.UnslidAddress:
		return formatInteger()
	case uint:
		return formatInteger()
	case uint8:
		return formatInteger()
	case uint16:
		return formatInteger()
	case uint32:
		return formatInteger()
	case uint64:
		return formatInteger()
	case int:
		return formatInteger()
	case int8:
		return formatInteger()
	case int16:
		return formatInteger()
	case int32:
		return formatInteger()
	case int64:
		return formatInteger()
	case [16]byte:
		// FIXME: no way to distinguish between []byte and []uint8
		if name == "UUID" {
			return fmt.Sprintf(
				"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
				v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], v[11], v[12], v[13], v[14], v[15],
			)
		}
		return commons.FromCString(v[:])
	case [32]byte:
		return commons.FromCString(v[:])
	}
	return fmt.Sprintf("%v", value)
}
