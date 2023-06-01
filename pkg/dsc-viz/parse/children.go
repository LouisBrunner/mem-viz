package parse

import (
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"golang.org/x/exp/slices"
)

func addChildDeep(parent, child *contracts.MemoryBlock) *contracts.MemoryBlock {
	for i, curr := range parent.Content {
		if parsingutils.IsInsideOf(child, curr) {
			return addChildDeep(curr, child)
		} else if curr.Address > child.Address {
			parent.Content = slices.Insert(parent.Content, i, child)
			return parent
		}
	}
	parent.Content = append(parent.Content, child)
	return parent
}

func addChildReal(parent, child *contracts.MemoryBlock) {
	parent.Content = append(parent.Content, child)
}

func (me *parser) addChildFast(block *contracts.MemoryBlock) {
	sameAddress, found := me.allBlocks[block.Address]
	if !found {
		sameAddress = &[]*contracts.MemoryBlock{}
		me.allBlocks[block.Address] = sameAddress
	}

	added := false
	for i, curr := range *sameAddress {
		if parsingutils.LessThan(block, curr) {
			*sameAddress = slices.Insert(*sameAddress, i, block)
			added = true
			break
		}
	}
	if !added {
		*sameAddress = append(*sameAddress, block)
	}
}

func (me *parser) removeChild(block *contracts.MemoryBlock) {
	sameAddresses, found := me.allBlocks[block.Address]
	if !found {
		return
	}
	index := slices.Index(*sameAddresses, block)
	if index < 0 {
		return
	}
	newAddresses := slices.Delete(*sameAddresses, index, index+1)
	me.allBlocks[block.Address] = &newAddresses
}
