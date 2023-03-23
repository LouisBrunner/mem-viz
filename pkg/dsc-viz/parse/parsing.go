package parse

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func parseAndAdd[T any](r io.Reader, parent *contracts.MemoryBlock, from *contracts.MemoryBlock, offset uint64, data T, label string) (*contracts.MemoryBlock, error) {
	err := commons.Unpack(r, &data)
	if err != nil {
		return nil, err
	}
	block := createStructBlock(parent, data, label, offset)
	return block, nil
}

func parseAndAddStruct[T any](cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset uint64, data T, label string) (*contracts.MemoryBlock, error) {
	if offset == 0 {
		return nil, nil
	}

	block, err := parseAndAdd(cache.ReaderAtOffset(int64(offset)), inside, from, from.ParentOffset+offset, data, label)
	if err != nil {
		return nil, err
	}
	err = addLink(from, fieldName, block, "points to")
	if err != nil {
		return nil, err
	}
	return block, nil
}

func parseAndAddMultipleStructs[T any](cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset uint64, countFieldName string, count uint64, data T, label string) ([]*contracts.MemoryBlock, error) {
	if offset == 0 {
		return nil, nil
	}

	offsetFromDefiner := from.ParentOffset + offset
	arrayBlock := createEmptyBlock(inside, fmt.Sprintf("%s (%d)", label, count), offsetFromDefiner)

	if count == 0 {
		return nil, nil
	}

	blocks := make([]*contracts.MemoryBlock, 0, count)
	for i := uint64(0); i < count; i += 1 {
		itemOffset := i * uint64(unsafe.Sizeof(data))
		block, err := parseAndAdd(cache.ReaderAtOffset(int64(offsetFromDefiner+itemOffset)), arrayBlock, from, itemOffset, data, fmt.Sprintf("%s %d/%d", label, i+1, count))
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	err := addLink(from, fieldName, blocks[0], "points to")
	if err != nil {
		return nil, err
	}
	err = addLink(from, countFieldName, blocks[0], "gives size")
	if err != nil {
		return nil, err
	}
	return blocks, nil
}
