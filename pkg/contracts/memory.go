package contracts

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
