package commons

import "github.com/LouisBrunner/mem-viz/pkg/contracts"

type VisitContext struct {
	Depth           int
	PreviousSibling *contracts.MemoryBlock
	NextSibling     *contracts.MemoryBlock
	Parent          *contracts.MemoryBlock
}

type BlockVisitor = func(ctx VisitContext, block *contracts.MemoryBlock) error
type ValueVisitor = func(ctx VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue) error
type LinkVisitor = func(ctx VisitContext, block *contracts.MemoryBlock, value *contracts.MemoryValue, link *contracts.MemoryLink) error

// FIXME: recursion everywhere... bad!

type VisitorSetup struct {
	BeforeChildren BlockVisitor
	AfterChildren  BlockVisitor
}

func visitEachBlockInternal(root *contracts.MemoryBlock, ctx VisitContext, visitor VisitorSetup) error {
	if visitor.BeforeChildren != nil {
		err := visitor.BeforeChildren(ctx, root)
		if err != nil {
			return err
		}
	}
	previousSibling := ctx.PreviousSibling
	for i, child := range root.Content {
		nextSibling := ctx.NextSibling
		if i < len(root.Content)-1 {
			nextSibling = root.Content[i+1]
		}
		err := visitEachBlockInternal(child, VisitContext{
			Depth:           ctx.Depth + 1,
			PreviousSibling: previousSibling,
			NextSibling:     nextSibling,
			Parent:          root,
		}, visitor)
		if err != nil {
			return err
		}
		previousSibling = child
	}
	if visitor.AfterChildren != nil {
		err := visitor.AfterChildren(ctx, root)
		if err != nil {
			return err
		}
	}
	return nil
}

func VisitEachBlockAdvanced(root *contracts.MemoryBlock, visitor VisitorSetup) error {
	return visitEachBlockInternal(root, VisitContext{
		Depth:           0,
		PreviousSibling: nil,
		Parent:          nil,
	}, visitor)
}

func VisitEachBlock(root *contracts.MemoryBlock, visitor BlockVisitor) error {
	return visitEachBlockInternal(root, VisitContext{
		Depth:           0,
		PreviousSibling: nil,
		Parent:          nil,
	}, VisitorSetup{BeforeChildren: visitor})
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
