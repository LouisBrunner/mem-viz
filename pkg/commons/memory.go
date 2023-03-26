package commons

import "github.com/LouisBrunner/mem-viz/pkg/contracts"

type VisitContext struct {
	Depth           int
	PreviousSibling *contracts.MemoryBlock
	Parent          *contracts.MemoryBlock
}

type BlockVisitor = func(ctx VisitContext, block *contracts.MemoryBlock) error
type ValueVisitor = func(ctx VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue) error
type LinkVisitor = func(ctx VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error

// FIXME: recursion everywhere... bad!

func visitEachBlockAux(root *contracts.MemoryBlock, ctx VisitContext, visitor BlockVisitor) error {
	err := visitor(ctx, root)
	if err != nil {
		return err
	}
	var previousSibling *contracts.MemoryBlock
	for _, child := range root.Content {
		err = visitEachBlockAux(child, VisitContext{
			Depth:           ctx.Depth + 1,
			PreviousSibling: previousSibling,
			Parent:          root,
		}, visitor)
		if err != nil {
			return err
		}
		previousSibling = child
	}
	return nil
}

func VisitEachBlock(root *contracts.MemoryBlock, visitor BlockVisitor) error {
	return visitEachBlockAux(root, VisitContext{
		Depth:           0,
		PreviousSibling: nil,
		Parent:          nil,
	}, visitor)
}

func VisitEachValue(root *contracts.MemoryBlock, visitor ValueVisitor) error {
	return VisitEachBlock(root, func(ctx VisitContext, block *contracts.MemoryBlock) error {
		for _, value := range block.Values {
			if err := visitor(ctx, block, value); err != nil {
				return err
			}
		}
		return nil
	})
}

func VisitEachLink(root *contracts.MemoryBlock, visitor LinkVisitor) error {
	return VisitEachValue(root, func(ctx VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue) error {
		for _, link := range value.Links {
			if err := visitor(ctx, block, value, link); err != nil {
				return err
			}
		}
		return nil
	})
}
