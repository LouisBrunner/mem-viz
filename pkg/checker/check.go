package checker

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

func blockDetails(block *contracts.MemoryBlock) string {
	return fmt.Sprintf("%q (%#016x-%#016x)", block.Name, block.Address, block.Address+uintptr(block.GetSize()))
}

func valueDetails(value *contracts.MemoryValue) string {
	return fmt.Sprintf("%q (%#04x-%#04x)", value.Name, value.Offset, value.Offset+uint64(value.Size))
}

func Check(logger *logrus.Logger, mb *contracts.MemoryBlock) error {
	return commons.VisitEachBlock(mb, func(ctx commons.VisitContext, block *contracts.MemoryBlock) error {
		size := uintptr(block.GetSize())
		parentEnd := block.Address + size

		previousAddress := uintptr(0)
		for i, child := range block.Content {
			childSize := uintptr(child.GetSize())
			childEnd := child.Address + childSize
			if child.Address < block.Address {
				return fmt.Errorf("child %v is out of bounds (before) of its parent %v", blockDetails(child), blockDetails(block))
			}
			if child.Address < previousAddress {
				return fmt.Errorf("children of %v are not sorted: %v should be after %v", blockDetails(block), blockDetails(child), blockDetails(block.Content[i-1]))
			}
			if block.Size != 0 && parentEnd < childEnd {
				return fmt.Errorf("child %v is out of bounds (after) of its parent %v", blockDetails(child), blockDetails(block))
			}
			if uintptr(child.ParentOffset) != child.Address-block.Address {
				return fmt.Errorf("child %v has an invalid offset: %#x != %#x", blockDetails(child), child.ParentOffset, child.Address-block.Address)
			}
			previousAddress = childEnd
		}

		previousOffset := uint64(0)
		for i, value := range block.Values {
			if value.Offset < previousOffset {
				return fmt.Errorf("values of %v are not sorted: %v should be after %v", blockDetails(block), valueDetails(value), valueDetails(block.Values[i-1]))
			}
			if uintptr(value.Offset)+uintptr(value.Size) > size {
				return fmt.Errorf("value %v is out of bounds of its parent %v", valueDetails(value), blockDetails(block))
			}
		}

		return nil
	})
}
