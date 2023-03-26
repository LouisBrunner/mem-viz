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

func (me *parser) addArrayBlock(frame *blockFrame, fieldName string, offset subcontracts.Address, countFieldName string, count uint64, data any, label string, create func(label string, offset subcontracts.Address, size uint64) (*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, subcontracts.Address, uint64, error) {
	if offset.Invalid() || count == 0 {
		return nil, nil, 0, nil
	}

	val := getDataValue(data)
	size := uint64(val.Type().Size())

	offsetFromDefiner := offset.AddBase(uintptr(frame.offsetFromStart + frame.parentStruct.ParentOffset))
	arrayBlock, err := create(fmt.Sprintf("%s (%d)", label, count), offsetFromDefiner, size*count)
	if err != nil {
		return nil, nil, 0, err
	}
	err = addLink(frame.parentStruct, fieldName, arrayBlock, "points to")
	if err != nil {
		return nil, nil, 0, err
	}
	if me.addSizeLink {
		err = addLink(frame.parentStruct, countFieldName, arrayBlock, "gives size")
		if err != nil {
			return nil, nil, 0, err
		}
	}
	return arrayBlock, offsetFromDefiner, size, nil
}

type arrayElement struct {
	Block *contracts.MemoryBlock
	Data  interface{}
}

func (me *parser) parseAndAddArray(frame *blockFrame, fieldName string, offset subcontracts.Address, countFieldName string, count uint64, data any, label string) (*contracts.MemoryBlock, []arrayElement, error) {
	if me.thresholdsArrayTooBig != 0 && count > me.thresholdsArrayTooBig {
		arrayBlock, err := me.parseAndAddArrayOnly(frame, fieldName, offset, countFieldName, count, data, label)
		return arrayBlock, nil, err
	}

	arrayBlock, offset, size, err := me.addArrayBlock(frame, fieldName, offset, countFieldName, count, data, label, func(label string, offset subcontracts.Address, size uint64) (*contracts.MemoryBlock, error) {
		return me.createEmptyBlock(frame.parent, label, offset)
	})
	if err != nil {
		return nil, nil, err
	}
	if arrayBlock == nil {
		return nil, nil, nil
	}

	items := make([]arrayElement, 0, count)
	for i := uint64(0); i < count; i += 1 {
		itemOffset := i * size
		block, err := me.parseAndAdd(offset.GetReader(frame.cache, itemOffset, me.slide), arrayBlock, frame.parentStruct, subcontracts.RelativeAddress64(itemOffset), data, fmt.Sprintf("%s %d/%d", label, i+1, count))
		if err != nil {
			return nil, nil, err
		}

		items = append(items, arrayElement{Block: block, Data: copyDataValue(data)})
	}
	return arrayBlock, items, nil
}

func (me *parser) parseAndAddArrayOnly(frame *blockFrame, fieldName string, offset subcontracts.Address, countFieldName string, count uint64, data any, label string) (*contracts.MemoryBlock, error) {
	arrayBlock, _, _, err := me.addArrayBlock(frame, fieldName, offset, countFieldName, count, data, label, func(label string, offset subcontracts.Address, size uint64) (*contracts.MemoryBlock, error) {
		return me.createCommonBlock(frame.parent, label, offset, size)
	})
	return arrayBlock, err
}

func (me *parser) parseAndAddBlob(frame *blockFrame, fieldName string, offset subcontracts.Address, fieldSizeName string, size uint64, data any, label string) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	blob, err := me.createBlobBlock(frame, fieldName, offset, fieldSizeName, size, label)
	if err != nil {
		return nil, nil, err
	}
	if blob == nil {
		return nil, nil, nil
	}
	header, err := me.parseAndAdd(offset.GetReader(frame.cache, 0, me.slide), blob, frame.parent, subcontracts.RelativeAddress32(0), data, fmt.Sprintf("%s Header", label))
	if err != nil {
		return nil, nil, nil
	}
	return blob, header, nil
}
