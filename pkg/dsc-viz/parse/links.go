package parse

import (
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
)

func (me *parser) addLinkWithOffset(frame *blockFrame, parentValueName string, offset subcontracts.Address, linkName string) error {
	if offset.Invalid() {
		return nil
	}

	return parsingutils.AddLinkWithAddr(frame.parentStruct, parentValueName, linkName, offset.AddBase(frame.parent.Address).Calculate(me.slide))
}
