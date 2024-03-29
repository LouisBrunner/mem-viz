package parse

import (
	"fmt"
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
)

func (me *parser) createCommonBlock(parent *contracts.MemoryBlock, label string, offset subcontracts.Address, size uint64) (*contracts.MemoryBlock, error) {
	address := offset.AddBase(parent.Address).Calculate(me.slide)
	block := &contracts.MemoryBlock{
		Name:         label,
		Address:      address,
		Size:         size,
		ParentOffset: uint64(address - parent.Address),
	}
	if parent == me.root || me.isUnique(parent) {
		addChildReal(parent, block)
	} else {
		me.addChildFast(block)
	}
	return block, nil
}

func (me *parser) createStructBlock(parent *contracts.MemoryBlock, data any, label string, offset subcontracts.Address) (*contracts.MemoryBlock, error) {
	val := parsingutils.GetDataValue(data)
	typ := val.Type()

	block, err := me.createCommonBlock(parent, label, offset, uint64(typ.Size()))
	if err != nil {
		return nil, err
	}

	if typ.Kind() != reflect.Struct {
		addValue(block, "Value", val.Interface(), 0, uint8(typ.Size()))
		return block, nil
	}

	for _, field := range reflect.VisibleFields(typ) {
		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct && field.Anonymous {
			continue
		}
		addValue(block, field.Name, val.FieldByIndex(field.Index).Interface(), uint64(field.Offset), uint8(fieldType.Size()))
	}

	return block, nil
}

func (me *parser) createBlobBlock(frame *blockFrame, fieldName string, offset subcontracts.Address, fieldSizeName string, size uint64, label string) (*contracts.MemoryBlock, error) {
	if offset.Invalid() || size == 0 {
		return nil, nil
	}

	block, err := me.createCommonBlock(frame.parent, label, offset, size)
	if err != nil {
		return nil, err
	}
	err = parsingutils.AddLinkWithBlock(frame.parentStruct, fieldName, block, "points to")
	if err != nil {
		return nil, err
	}
	if me.addSizeLink {
		err = parsingutils.AddLinkWithBlock(frame.parentStruct, fieldSizeName, block, "gives size")
		if err != nil {
			return nil, err
		}
	}
	return block, nil
}

func (me *parser) createEmptyBlock(parent *contracts.MemoryBlock, label string, offset subcontracts.Address) (*contracts.MemoryBlock, error) {
	return me.createCommonBlock(parent, label, offset, 0)
}

func (me *parser) createArrayBlock(frame *blockFrame, fieldName string, offset subcontracts.Address, countFieldName string, count uint64, data any, label string, create func(label string, offset subcontracts.Address, size uint64) (*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, subcontracts.Address, uint64, error) {
	if offset.Invalid() || count == 0 {
		return nil, nil, 0, nil
	}

	val := parsingutils.GetDataValue(data)
	size := uint64(val.Type().Size())

	offsetFromDefiner := offset.AddBase(uintptr(frame.offsetFromStart + frame.parentStruct.ParentOffset))
	arrayBlock, err := create(fmt.Sprintf("%s (%d)", label, count), offsetFromDefiner, size*count)
	if err != nil {
		return nil, nil, 0, err
	}
	err = parsingutils.AddLinkWithBlock(frame.parentStruct, fieldName, arrayBlock, "points to")
	if err != nil {
		return nil, nil, 0, err
	}
	if me.addSizeLink {
		err = parsingutils.AddLinkWithBlock(frame.parentStruct, countFieldName, arrayBlock, "gives size")
		if err != nil {
			return nil, nil, 0, err
		}
	}
	return arrayBlock, offsetFromDefiner, size, nil
}
