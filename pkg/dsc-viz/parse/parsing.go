package parse

import (
	"fmt"
	"io"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseAndAdd(r io.Reader, parent *contracts.MemoryBlock, from *contracts.MemoryBlock, offset subcontracts.Address, data any, label string) (*contracts.MemoryBlock, error) {
	err := commons.Unpack(r, data)
	if err != nil {
		return nil, err
	}
	return me.createStructBlock(parent, data, label, offset)
}

func (me *parser) parseAndAddArray(cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset subcontracts.Address, countFieldName string, count uint64, data any, label string) ([]*contracts.MemoryBlock, error) {
	if offset.Invalid() {
		return nil, nil
	}

	offsetFromDefiner := offset.AddBase(uintptr(from.ParentOffset))
	arrayBlock, err := me.createEmptyBlock(inside, fmt.Sprintf("%s (%d)", label, count), offsetFromDefiner)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	val := getDataValue(data)

	blocks := make([]*contracts.MemoryBlock, 0, count)
	for i := uint64(0); i < count; i += 1 {
		itemOffset := i * uint64(val.Type().Size())
		block, err := me.parseAndAdd(offsetFromDefiner.GetReader(cache, itemOffset, me.slide), arrayBlock, from, subcontracts.RelativeAddress64(itemOffset), data, fmt.Sprintf("%s %d/%d", label, i+1, count))
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

func (me *parser) parseAndAddBlob(cache subcontracts.Cache, inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset subcontracts.Address, fieldSizeName string, size uint64, data any, label string) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	blob, err := me.createBlobBlock(inside, from, fieldName, offset, fieldSizeName, size, label)
	if err != nil {
		return nil, nil, err
	}
	if blob == nil {
		return nil, nil, nil
	}
	header, err := me.parseAndAdd(offset.GetReader(cache, 0, me.slide), blob, from, subcontracts.RelativeAddress32(0), data, fmt.Sprintf("%s Header", label))
	if err != nil {
		return nil, nil, nil
	}
	return blob, header, nil
}
