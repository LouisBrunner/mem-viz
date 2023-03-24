package commons

import "github.com/LouisBrunner/mem-viz/pkg/contracts"

type BlockVisitor = func(depth int, block *contracts.MemoryBlock) error
type ValueVisitor = func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue) error
type LinkVisitor = func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error

// FIXME: recursion everywhere... bad!

func visitEachBlockAux(root *contracts.MemoryBlock, depth int, visitor BlockVisitor) error {
	if err := visitor(depth, root); err != nil {
		return err
	}
	for _, child := range root.Content {
		if err := visitEachBlockAux(child, depth+1, visitor); err != nil {
			return err
		}
	}
	return nil
}

func VisitEachBlock(root *contracts.MemoryBlock, visitor BlockVisitor) error {
	return visitEachBlockAux(root, 0, visitor)
}

func VisitEachValue(root *contracts.MemoryBlock, visitor ValueVisitor) error {
	return VisitEachBlock(root, func(depth int, block *contracts.MemoryBlock) error {
		for _, value := range block.Values {
			if err := visitor(depth+1, block, value); err != nil {
				return err
			}
		}
		return nil
	})
}

func VisitEachLink(root *contracts.MemoryBlock, visitor LinkVisitor) error {
	return VisitEachValue(root, func(depth int, block *contracts.MemoryBlock, value *contracts.MemoryValue) error {
		for _, link := range value.Links {
			if err := visitor(depth+1, block, value, link); err != nil {
				return err
			}
		}
		return nil
	})
}
