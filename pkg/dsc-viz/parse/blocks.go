package parse

import (
	"reflect"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func createStructBlock[T any, A addressOrOffset](parent *contracts.MemoryBlock, data T, label string, offset A) (*contracts.MemoryBlock, error) {
	empty := interface{}(data) == nil
	size := uint64(0)
	if !empty {
		size = uint64(unsafe.Sizeof(data))
	}
	block, err := createCommonBlock(parent, label, offset, size)
	if err != nil {
		return nil, err
	}

	if empty {
		return block, nil
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

	return block, nil
}

func createBlobBlock[A addressOrOffset](inside *contracts.MemoryBlock, from *contracts.MemoryBlock, fieldName string, offset A, fieldSizeName string, size uint64, label string) (*contracts.MemoryBlock, error) {
	if offset == 0 || size == 0 {
		return nil, nil
	}

	block, err := createCommonBlock(inside, label, offset, size)
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

func createEmptyBlock[A addressOrOffset](parent *contracts.MemoryBlock, label string, offset A) (*contracts.MemoryBlock, error) {
	return createCommonBlock(parent, label, offset, 0)
}
