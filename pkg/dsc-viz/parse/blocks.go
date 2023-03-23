package parse

import (
	"reflect"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func createStructBlock[T any](parent *contracts.MemoryBlock, data T, label string, offset uint64) *contracts.MemoryBlock {
	empty := interface{}(data) == nil
	size := uint64(0)
	if !empty {
		size = uint64(unsafe.Sizeof(data))
	}
	block := createCommonBlock(parent, label, offset, size)

	if empty {
		return block
	}

	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
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

	return block
}

func createBlobBlock(inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset uint64, fieldSizeName string, size uint64, label string) (*contracts.MemoryBlock, error) {
	if offset == 0 || size == 0 {
		return nil, nil
	}

	block := createCommonBlock(inside, label, offset, size)
	err := addLink(from, fieldName, block, "points to")
	if err != nil {
		return nil, err
	}
	err = addLink(from, fieldSizeName, block, "gives size")
	if err != nil {
		return nil, err
	}
	return block, nil
}

func createEmptyBlock(parent *contracts.MemoryBlock, label string, offset uint64) *contracts.MemoryBlock {
	return createCommonBlock(parent, label, offset, 0)
}
