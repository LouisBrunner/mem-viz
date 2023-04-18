package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func addLinkCommon(parent *contracts.MemoryBlock, parentValueName, linkName string, addr uintptr) error {
	if parentValueName == "" {
		return nil
	}

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
