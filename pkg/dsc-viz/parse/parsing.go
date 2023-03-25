package parse

import (
	"fmt"
	"io"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

type addressOrOffset interface {
	uint64 | subcontracts.UnslidAddress
}

func parseAndAdd[T any, A addressOrOffset](r io.Reader, parent *contracts.MemoryBlock, from *contracts.MemoryBlock, offset A, data T, label string) (*contracts.MemoryBlock, error) {
	err := commons.Unpack(r, &data)
	if err != nil {
		return nil, err
	}
	return createStructBlock(parent, data, label, offset)
}

func parseAndAddMultipleStructs[T any, A addressOrOffset](cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset A, countFieldName string, count uint64, data T, label string) ([]*contracts.MemoryBlock, error) {
	if offset == 0 {
		return nil, nil
	}

	offsetFromDefiner := addParentOffset(offset, from.ParentOffset)
	arrayBlock, err := createEmptyBlock(inside, fmt.Sprintf("%s (%d)", label, count), offsetFromDefiner)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	blocks := make([]*contracts.MemoryBlock, 0, count)
	for i := uint64(0); i < count; i += 1 {
		itemOffset := i * uint64(unsafe.Sizeof(data))
		block, err := parseAndAdd(getReaderAtOffset(cache, offsetFromDefiner, itemOffset), arrayBlock, from, itemOffset, data, fmt.Sprintf("%s %d/%d", label, i+1, count))
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	err = addLink(from, fieldName, blocks[0], "points to")
	if err != nil {
		return nil, err
	}
	err = addLink(from, countFieldName, blocks[0], "gives size")
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func parseAndAddBlob[T any, A addressOrOffset](cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset A, fieldSizeName string, size uint64, data T, label string) (*contracts.MemoryBlock, *contracts.MemoryBlock, T, error) {
	blob, err := createBlobBlock(inside, from, fieldName, offset, fieldSizeName, size, label)
	if err != nil {
		return nil, nil, data, err
	}
	if blob == nil {
		return nil, nil, data, nil
	}
	header, err := parseAndAdd(getReaderAtOffset(cache, offset, 0), blob, from, uint64(0), data, fmt.Sprintf("%s Header", label))
	if err != nil {
		return nil, nil, data, nil
	}
	return blob, header, data, nil
}
