package parse

import (
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

type blockFrame struct {
	cache           subcontracts.Cache
	parent          *contracts.MemoryBlock
	parentStruct    *contracts.MemoryBlock
	offsetFromStart uint64
}

func topFrame(cache subcontracts.Cache, parent, parentStruct *contracts.MemoryBlock) *blockFrame {
	return &blockFrame{
		cache:           cache,
		parent:          parent,
		parentStruct:    parentStruct,
		offsetFromStart: 0,
	}
}

func (f *blockFrame) pushFrame(parent, parentStruct *contracts.MemoryBlock) *blockFrame {
	return &blockFrame{
		cache:           f.cache,
		parent:          parent,
		parentStruct:    parentStruct,
		offsetFromStart: f.offsetFromStart + f.parent.ParentOffset,
	}
}

func (f *blockFrame) siblingFrame(parentStruct *contracts.MemoryBlock) *blockFrame {
	return &blockFrame{
		cache:           f.cache,
		parent:          f.parent,
		parentStruct:    parentStruct,
		offsetFromStart: f.offsetFromStart,
	}
}
