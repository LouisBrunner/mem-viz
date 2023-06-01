package parsingutils

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func AddLinkWithAddr(parent *contracts.MemoryBlock, parentValueName, linkName string, addr uintptr) error {
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

func AddLinkWithBlock(parent *contracts.MemoryBlock, parentValueName string, child *contracts.MemoryBlock, linkName string) error {
	return AddLinkWithAddr(parent, parentValueName, linkName, child.Address)
}
