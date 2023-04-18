package parse

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
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
	case subcontracts.UnslidAddress32:
		return formatInteger()
	case subcontracts.RelativeAddress32:
		return formatInteger()
	case subcontracts.RelativeAddress64:
		return formatInteger()
	case subcontracts.LinkEditOffset:
		return formatInteger()
	case subcontracts.ManualAddress:
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
	case []uint16:
		return fmt.Sprintf("[%d]uint16", len(v))
	case fmt.Stringer:
		return v.String()
	}
	jsond, err := json.Marshal(value)
	if err == nil {
		return string(jsond)
	}
	return fmt.Sprintf("%v", value)
}
