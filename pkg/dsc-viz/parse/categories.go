package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

type category int

const (
	categoryPaths        category = iota
	categoryImages       category = iota
	categoryMappings     category = iota
	categoryLinkEdit     category = iota
	categoryPostLinkEdit category = iota
)

var allCategories = []category{
	categoryPaths,
	categoryImages,
	categoryMappings,
	categoryLinkEdit,
}

func (me category) IsGlobal() bool {
	switch me {
	case categoryImages:
		fallthrough
	case categoryLinkEdit:
		fallthrough
	case categoryPostLinkEdit:
		return true
	}
	return false
}

func (me *parser) clearNonGlobalCategories() {
	for _, cat := range allCategories {
		if !cat.IsGlobal() {
			me.uniqueBlocks[cat] = nil
		}
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
	me.uniqueBlocks[cat] = append(me.uniqueBlocks[cat], newBlock)
	return newBlock, nil
}

func (me *parser) updateCategoryBlock(block *contracts.MemoryBlock, baseName string) {
	block.Name = fmt.Sprintf("%s (%d)", baseName, len(block.Content))
	block.Size = 0
	block.Size = block.GetSize()
}
