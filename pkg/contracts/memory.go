package contracts

type MemoryBlock struct {
	Name         string
	Address      uintptr
	Size         uint64
	ParentOffset uint64
	Content      []MemoryBlock
	Values       []MemoryValue
}

type MemoryValue struct {
	Name   string
	Offset uint64
	Value  string
	Links  []MemoryLink
}

type MemoryLink struct {
	Name          string
	TargetAddress uint64
}
