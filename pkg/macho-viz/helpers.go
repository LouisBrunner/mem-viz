package macho

import (
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"golang.org/x/exp/slices"
)

func (me *parser) addChildDeep(parent, child *contracts.MemoryBlock) *contracts.MemoryBlock {
	for i, curr := range parent.Content {
		if parsingutils.IsInsideOf(child, curr) {
			return me.addChildDeep(curr, child)
		} else if curr.Address > child.Address {
			parent.Content = slices.Insert(parent.Content, i, child)
			return parent
		}
	}
	parent.Content = append(parent.Content, child)
	return parent
}

func (me *parser) addChild(parent, child *contracts.MemoryBlock) *contracts.MemoryBlock {
	sameAddress, found := me.allBlocks[child.Address]
	if !found {
		sameAddress = &[]*contracts.MemoryBlock{}
		me.allBlocks[child.Address] = sameAddress
	}

	added := false
	for i, curr := range *sameAddress {
		if parsingutils.LessThan(child, curr) {
			*sameAddress = slices.Insert(*sameAddress, i, child)
			added = true
			break
		}
	}
	if !added {
		*sameAddress = append(*sameAddress, child)
	}
	return child
}

func (me *parser) addStructDetailed(parent *contracts.MemoryBlock, data interface{}, name string, offset, size uint64, banned []string) *contracts.MemoryBlock {
	val := parsingutils.GetDataValue(data)
	typ := val.Type()

	if size == 0 {
		size = uint64(typ.Size())
	}
	block := me.addChild(parent, &contracts.MemoryBlock{
		Name:         name,
		Address:      parent.Address + uintptr(offset),
		ParentOffset: offset,
		Size:         size,
	})

	if typ.Kind() != reflect.Struct {
		parsingutils.AddValue(block, "Value", val.Interface(), 0, uint8(typ.Size()), parsingutils.FormatValue)
		return block
	}

	fieldOffset := uint64(0)
	for _, field := range reflect.VisibleFields(typ) {
		fieldType := field.Type
		fieldVal := val.FieldByIndex(field.Index)
		if (fieldType.Kind() == reflect.Struct && field.Anonymous) || !field.IsExported() || slices.Contains(banned, field.Name) {
			continue
		}
		size := uint8(fieldType.Size())
		if fieldType.Kind() == reflect.Slice {
			size = uint8(int(fieldType.Elem().Size()) * fieldVal.Len())
		}
		parsingutils.AddValue(block, field.Name, fieldVal.Interface(), fieldOffset, size, parsingutils.FormatValue)
		fieldOffset += uint64(size)
	}

	return block
}

func (me *parser) addStruct(parent *contracts.MemoryBlock, data interface{}, name string, offset uint64) *contracts.MemoryBlock {
	return me.addStructDetailed(parent, data, name, offset, 0, nil)
}
