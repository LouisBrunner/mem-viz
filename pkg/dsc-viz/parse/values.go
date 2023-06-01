package parse

import (
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
)

// FIXME: disgusting reflection but no other way unless Go supports generics on function receivers

func addValue(block *contracts.MemoryBlock, name string, value interface{}, offset uint64, size uint8) {
	parsingutils.AddValue(block, name, value, offset, size, func(name string, value interface{}) string {
		switch value.(type) {
		case subcontracts.UnslidAddress, subcontracts.UnslidAddress32, subcontracts.RelativeAddress32, subcontracts.RelativeAddress64, subcontracts.LinkEditOffset, subcontracts.ManualAddress:
			return parsingutils.FormatInteger(name, value)
		default:
			return parsingutils.FormatValue(name, value)
		}
	})
}

func copyDataValue(v interface{}) interface{} {
	dat := reflect.ValueOf(v)
	for dat.Kind() == reflect.Ptr {
		dat = dat.Elem()
	}
	return dat.Interface()
}
