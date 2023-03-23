package parse

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"golang.org/x/exp/slices"
)

func addLink(parent *contracts.MemoryBlock, parentValueName string, child *contracts.MemoryBlock, linkName string) error {
	for i := range parent.Values {
		if parent.Values[i].Name != parentValueName {
			continue
		}

		parent.Values[i].Links = append(parent.Values[i].Links, &contracts.MemoryLink{
			Name:          linkName,
			TargetAddress: uint64(child.Address),
		})
		return nil
	}

	return fmt.Errorf("could not find value %q in parent %+v", parentValueName, parent)
}

func createBlock[T any](parent *contracts.MemoryBlock, data T, label string, offset uint64) *contracts.MemoryBlock {
	headerBlock := &contracts.MemoryBlock{
		Name:         label,
		Address:      parent.Address + uintptr(offset),
		Size:         uint64(unsafe.Sizeof(data)),
		ParentOffset: offset,
	}

	val := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
	for _, field := range reflect.VisibleFields(typ) {
		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct && field.Anonymous {
			continue
		}

		headerBlock.Values = append(headerBlock.Values, &contracts.MemoryValue{
			Name:   field.Name,
			Value:  formatValue(field.Name, val.FieldByIndex(field.Index).Interface()),
			Offset: uint64(field.Offset),
			Size:   uint8(fieldType.Size()),
		})
	}

	addChild(parent, headerBlock)
	return headerBlock
}

func formatValue(name string, value interface{}) string {
	formatInteger := func() string {
		// FIXME: would be nice to have more specialized formats, e.g. for OSVersion or CacheType
		return fmt.Sprintf("%d (%#x)", value, value)
	}

	switch v := value.(type) {
	case uint:
		return formatInteger()
	case uint8:
		return formatInteger()
	case uint16:
		return formatInteger()
	case uint32:
		return formatInteger()
	case uint64:
		return formatInteger()
	case int:
		return formatInteger()
	case int8:
		return formatInteger()
	case int16:
		return formatInteger()
	case int32:
		return formatInteger()
	case int64:
		return formatInteger()
	case [16]byte:
		// FIXME: no way to distinguish between []byte and []uint8
		if name == "UUID" {
			return fmt.Sprintf(
				"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
				v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], v[11], v[12], v[13], v[14], v[15],
			)
		}
		return strings.TrimRight(string(v[:]), "\x00") // FIXME: surely there is a way to build a utils for this?
	case [32]byte:
		return strings.TrimRight(string(v[:]), "\x00")
	}
	return fmt.Sprintf("%v", value)
}

func addChild(parent, child *contracts.MemoryBlock) {
	for i, curr := range parent.Content {
		if curr.Address > child.Address {
			parent.Content = slices.Insert(parent.Content, i, child)
			return
		}
	}
	parent.Content = append(parent.Content, child)
}
