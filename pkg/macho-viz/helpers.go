package macho

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"golang.org/x/exp/slices"
)

func isOnEdge(child, parent *contracts.MemoryBlock) bool {
	return child.Address == parent.Address || child.Address == parent.Address+uintptr(parent.GetSize())
}

func (me *parser) addChildDeep(parent, child *contracts.MemoryBlock) *contracts.MemoryBlock {
	isEmpty := child.GetSize() == 0
	for i, curr := range parent.Content {
		if parsingutils.IsInsideOf(child, curr) && !(isEmpty && isOnEdge(child, curr)) {
			return me.addChildDeep(curr, child)
		} else if curr.Address > child.Address || (isEmpty && curr.Address == child.Address) {
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
	block := &contracts.MemoryBlock{
		Name:         name,
		Address:      parent.Address + uintptr(offset),
		ParentOffset: offset,
		Size:         size,
	}

	if typ.Kind() != reflect.Struct {
		addValue(block, "Value", val.Interface(), 0, uint8(size))
	} else {
		fieldOffset := uint64(0)
		fieldSize := uint64(0)
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
			addValue(block, field.Name, fieldVal.Interface(), fieldOffset, size)
			fieldOffset += uint64(size)
			fieldSize += uint64(size)
		}
		block.Size = fieldSize
	}

	return me.addChild(parent, block)
}

func (me *parser) addStruct(parent *contracts.MemoryBlock, data interface{}, name string, offset uint64) *contracts.MemoryBlock {
	return me.addStructDetailed(parent, data, name, offset, 0, nil)
}

var arm64ThreadEntries = []string{
	"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9",
	"x10", "x11", "x12", "x13", "x14", "x15", "x16", "x17", "x18", "x19",
	"x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28",
	"fp", "lr", "sp", "pc", "cpsr",
}
var amd64ThreadEntries = []string{
	"rax", "rbx", "rcx", "rdx", "rdi", "rsi", "rbp", "rsp", "r8", "r9",
	"r10", "r11", "r12", "r13", "r14", "r15", "rip", "rflags", "cs",
	"fs", "gs",
}
var flavouredThreadEntries = [][]string{
	arm64ThreadEntries,
	amd64ThreadEntries,
}

func addValue(parent *contracts.MemoryBlock, name string, value interface{}, offset uint64, size uint8) {
	parsingutils.AddValue(parent, name, value, offset, size, func(name string, value interface{}) string {
		switch v := value.(type) {
		case []byte: // FIXME: 64bit Thread Data assumption
			if len(v)%8 != 0 {
				return parsingutils.FormatValue(name, value)
			}
			first := true
			b := strings.Builder{}
			b.WriteString("{")
			for i := 0; i < len(v)/8; i++ {
				v := binary.LittleEndian.Uint64(v[i*8 : (i+1)*8])
				if v == 0 {
					continue
				}
				if !first {
					b.WriteString("; ")
				}
				labels := []string{}
				for _, flavour := range flavouredThreadEntries {
					if len(flavour) > i {
						labels = append(labels, flavour[i])
					}
				}
				label := strings.Join(labels, "/")
				if label == "" {
					label = "???"
				}
				b.WriteString(fmt.Sprintf("[%02d/%s]=%#012x", i, label, v))
				first = false
			}
			b.WriteString("}")
			return b.String()
		default:
			return parsingutils.FormatValue(name, value)
		}
	})
}
