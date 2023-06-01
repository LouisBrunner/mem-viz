package parsingutils

import (
	"strings"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func LessThan(a, b *contracts.MemoryBlock) bool {
	criteria := []int{
		b2i(a.Size == 0) - b2i(b.Size == 0),
		-int(a.Size - b.Size),
		len(a.Values) - len(b.Values),
		strings.Compare(a.Name, b.Name),
	}
	for _, c := range criteria {
		if c == 0 {
			continue
		} else if c < 0 {
			return true
		} else {
			return false
		}
	}
	return false
}

func IsInsideOf(child, parent *contracts.MemoryBlock) bool {
	return parent.Address <= child.Address && child.Address+uintptr(child.GetSize()) <= parent.Address+uintptr(parent.GetSize())
}
