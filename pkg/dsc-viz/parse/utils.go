package parse

import (
	"fmt"
	"io"
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"golang.org/x/exp/slices"
)

// FIXME: disgusting reflection but no other way unless Go supports generics on function receivers

func getDataValue(v interface{}) reflect.Value {
	dat := reflect.ValueOf(v)
	for dat.Kind() == reflect.Ptr {
		dat = dat.Elem()
	}
	return dat
}

func copyDataValue(v interface{}) interface{} {
	dat := reflect.ValueOf(v)
	for dat.Kind() == reflect.Ptr {
		dat = dat.Elem()
	}
	return dat.Interface()
}

func readCString(r io.Reader) string {
	var b []byte
	one := make([]byte, 1)
	for {
		c, err := r.Read(one)
		if err != nil || c == 0 || one[0] == 0 {
			break
		}
		b = append(b, one[:c]...)
	}
	return string(b)
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

func findAndRemoveChild(parent, child *contracts.MemoryBlock) error {
	for i, curr := range parent.Content {
		if curr == child {
			removeChild(parent, i)
			return nil
		}
	}
	return fmt.Errorf("could not find child %+v in parent %+v", child, parent)
}

func removeChild(parent *contracts.MemoryBlock, childIndex int) *contracts.MemoryBlock {
	child := parent.Content[childIndex]
	parent.Content = slices.Delete(parent.Content, childIndex, childIndex+1)
	return child
}

func moveChild(parent, newParent *contracts.MemoryBlock, childIndex int) {
	child := removeChild(parent, childIndex)
	child.ParentOffset = uint64(child.Address - newParent.Address)
	addChild(newParent, child)
}

func addLinkCommon(parent *contracts.MemoryBlock, parentValueName, linkName string, addr uintptr) error {
	for i := range parent.Values {
		if parent.Values[i].Name != parentValueName {
			continue
		}

		parent.Values[i].Links = append(parent.Values[i].Links, &contracts.MemoryLink{
			Name:          linkName,
			TargetAddress: uint64(addr),
		})
		return nil
	}

	return fmt.Errorf("could not find value %q in parent %+v", parentValueName, parent)
}

func addLink(parent *contracts.MemoryBlock, parentValueName string, child *contracts.MemoryBlock, linkName string) error {
	return addLinkCommon(parent, parentValueName, linkName, child.Address)
}

func (me *parser) addLinkWithOffset(frame *blockFrame, parentValueName string, offset subcontracts.Address, linkName string) error {
	if offset.Invalid() {
		return nil
	}

	return addLinkCommon(frame.parentStruct, parentValueName, linkName, offset.AddBase(frame.parent.Address).Calculate(me.slide))
}

func addValue(block *contracts.MemoryBlock, name string, value interface{}, offset uint64, size uint8) {
	block.Values = append(block.Values, &contracts.MemoryValue{
		Name:   name,
		Value:  formatValue(name, value),
		Offset: offset,
		Size:   size,
	})
}

func formatValue(name string, value interface{}) string {
	formatInteger := func() string {
		// FIXME: would be nice to have more specialized formats, e.g. for OSVersion or CacheType
		return fmt.Sprintf("%#x", value)
	}

	switch v := value.(type) {
	case subcontracts.UnslidAddress:
		return formatInteger()
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
		return commons.FromCString(v[:])
	case [32]byte:
		return commons.FromCString(v[:])
	}
	return fmt.Sprintf("%v", value)
}
