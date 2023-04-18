package parse

import (
	"io"
	"strings"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func readCString(r io.Reader) string {
	var b []byte
	one := make([]byte, 1)
	for {
		c, err := r.Read(one)
		if err != nil || c == 0 || one[0] == 0 {
			break
		}
		b = append(b, one[:c]...)
	}
	return string(b)
}

func isInsideOf(child, parent *contracts.MemoryBlock) bool {
	return parent.Address <= child.Address && child.Address+uintptr(child.GetSize()) <= parent.Address+uintptr(parent.GetSize())
}

func roundUp(x, y uint64) uint64 {
	return (x + y - 1) / y * y
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func lessThan(a, b *contracts.MemoryBlock) bool {
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
