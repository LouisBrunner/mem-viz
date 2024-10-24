//go:build !arm64

package macho

import (
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/blacktop/go-macho/types"
)

func (me *parser) addArchSpecificSection(parent *contracts.MemoryBlock, sect *types.Section) error {
	return nil
}
