package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"golang.org/x/exp/slices"
)

type category int

const (
	categoryPaths    category = iota
	categoryImages   category = iota
	categoryMappings category = iota
	categoryLinkEdit category = iota
	categoryStrings  category = iota
)

var allCategories = []category{
	categoryPaths,
	categoryImages,
	categoryMappings,
	categoryLinkEdit,
	categoryStrings,
}

func (me category) isGlobal() bool {
	switch me {
	case categoryImages:
		fallthrough
	case categoryLinkEdit:
		fallthrough
	case categoryStrings:
		return true
	}
	return false
}

func (me *parser) emptyCategory(cat category) {
	for _, curr := range me.uniqueBlocks[cat] {
		for _, child := range curr.Content {
			me.addChildFast(child)
		}
		curr.Content = nil
		me.addChildFast(curr)
	}
	me.uniqueBlocks[cat] = nil
}

func (me *parser) clearNonGlobalCategories() {
	for _, cat := range allCategories {
		if !cat.isGlobal() {
			me.emptyCategory(cat)
		}
	}
}

func (me *parser) flushCategories() {
	for _, cat := range allCategories {
		me.emptyCategory(cat)
	}
}

func (me *parser) findOrCreateUniqueBlock(cat category, finder func(i int, block *contracts.MemoryBlock) bool, creator func() (*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, error) {
	for i, block := range me.uniqueBlocks[cat] {
		if finder(i, block) {
			return block, nil
		}
	}
	newBlock, err := creator()
	if err != nil {
		return nil, err
	}
	me.removeChild(newBlock)
	me.uniqueBlocks[cat] = append(me.uniqueBlocks[cat], newBlock)
	return newBlock, nil
}

func (me *parser) updateCategoryBlock(block *contracts.MemoryBlock, baseName string) {
	block.Name = fmt.Sprintf("%s (%d)", baseName, len(block.Content))
	block.Size = 0
	block.Size = block.GetSize()
}

func (me *parser) isUnique(block *contracts.MemoryBlock) bool {
	for _, curr := range me.uniqueBlocks {
		if slices.Contains(curr, block) {
			return true
		}
	}
	return false
}
