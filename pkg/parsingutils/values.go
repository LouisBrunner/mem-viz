package parsingutils

import (
	"encoding/json"
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

type Formatter func(name string, value interface{}) string

func AddValue(block *contracts.MemoryBlock, name string, value interface{}, offset uint64, size uint8, format Formatter) {
	block.Values = append(block.Values, &contracts.MemoryValue{
		Name:   name,
		Value:  format(name, value),
		Offset: offset,
		Size:   size,
	})
}

func FormatInteger(name string, value interface{}) string {
	// FIXME: would be nice to have more specialized formats, e.g. for OSVersion or CacheType
	return fmt.Sprintf("%#x", value)
}

func FormatValue(name string, value interface{}) string {
	switch v := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, uintptr:
		return FormatInteger(name, v)
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
