package parse

import (
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func getDataValue(v interface{}) reflect.Value {
	dat := reflect.ValueOf(v)
	for dat.Kind() == reflect.Ptr {
		dat = dat.Elem()
	}
	return dat
}

func (me *parser) createStructBlock(parent *contracts.MemoryBlock, data any, label string, offset subcontracts.Address) (*contracts.MemoryBlock, error) {
	val := getDataValue(data)
	typ := val.Type()

	block, err := me.createCommonBlock(parent, label, offset, uint64(typ.Size()))
	if err != nil {
		return nil, err
	}

	for _, field := range reflect.VisibleFields(typ) {
		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct && field.Anonymous {
			continue
		}

		block.Values = append(block.Values, &contracts.MemoryValue{
			Name:   field.Name,
			Value:  formatValue(field.Name, val.FieldByIndex(field.Index).Interface()),
			Offset: uint64(field.Offset),
			Size:   uint8(fieldType.Size()),
		})
	}

	return block, nil
}

func (me *parser) createBlobBlock(inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset subcontracts.Address, fieldSizeName string, size uint64, label string) (*contracts.MemoryBlock, error) {
	if offset.Invalid() || size == 0 {
		return nil, nil
	}

	block, err := me.createCommonBlock(inside, label, offset, size)
	if err != nil {
		return nil, err
	}
	err = addLink(from, fieldName, block, "points to")
	if err != nil {
		return nil, err
	}
	err = addLink(from, fieldSizeName, block, "gives size")
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (me *parser) createEmptyBlock(parent *contracts.MemoryBlock, label string, offset subcontracts.Address) (*contracts.MemoryBlock, error) {
	return me.createCommonBlock(parent, label, offset, 0)
}
