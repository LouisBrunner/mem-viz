package contracts

// Expected guarantiees:
// - MemoryBlock.Content is ordered by Address (ASC)
// - MemoryBlock.Values is ordered by Offset (ASC)
// - MemoryValue.Links might be unordered
// - All Address are absolute
// - Parent.Address + MemoryBlock.ParentOffset == MemoryBlock.Address
// - Size can be 0 (always use MemoryBlock.GetSize)
// - If it isn't, MemoryBlock.Content should all fit inside the MemoryBlock (MemoryBlock.Address, MemoryBlock.Address + MemoryBlock.Size)

type MemoryBlock struct {
	Name         string
	Address      uintptr
	Size         uint64
	ParentOffset uint64
	Content      []*MemoryBlock
	Values       []*MemoryValue
}

func (m MemoryBlock) GetSize() uint64 {
	if m.Size != 0 {
		return m.Size
	}

	if len(m.Content) == 0 {
		return 0
	}

	last := m.Content[len(m.Content)-1]
	return uint64((last.Address + uintptr(last.GetSize())) - m.Address)
}

type MemoryValue struct {
	Name   string
	Offset uint64
	Size   uint8
	Value  string
	Links  []*MemoryLink
}

type MemoryLink struct {
	Name          string
	TargetAddress uint64
}
