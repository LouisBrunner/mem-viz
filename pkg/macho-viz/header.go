package macho

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
	"github.com/blacktop/go-macho"
)

func (me *parser) addHeader(root *contracts.MemoryBlock, m *macho.File) error {
	hdr := me.addStruct(root, m.FileHeader, "Header", 0)
	commands := me.addChild(root, &contracts.MemoryBlock{
		Name:         fmt.Sprintf("Commands (%d)", m.NCommands),
		Address:      root.Address + uintptr(hdr.Size),
		ParentOffset: uint64(hdr.Size),
		Size:         uint64(m.SizeCommands),
	})
	err := parsingutils.AddLinkWithBlock(hdr, "NCommands", commands, "gives amount")
	if err != nil {
		return err
	}
	err = parsingutils.AddLinkWithBlock(hdr, "SizeCommands", commands, "gives size")
	if err != nil {
		return err
	}

	offset := uint64(0)
	data := contextData{}
	for i, cmd := range m.Loads {
		block, err := me.addCommand(root, commands, i, cmd, offset, &data)
		if err != nil {
			return err
		}
		offset += uint64(block.Size)
	}
	return nil
}
