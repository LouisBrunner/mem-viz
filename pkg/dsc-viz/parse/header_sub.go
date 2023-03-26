package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) addSubCacheEntry(parent, headerBlock, subCache *contracts.MemoryBlock, header subcontracts.DYLDCacheHeaderV3, v2 *subcontracts.DYLDSubcacheEntryV2, v1 *subcontracts.DYLDSubcacheEntryV1, index uint64) error {
	var block *contracts.MemoryBlock
	var err error
	label := "Subcache Entry"
	if v2 != nil {
		block, err = me.createStructBlock(parent, *v2, fmt.Sprintf("%s (V2)", label), subcontracts.RelativeAddress64(index*uint64(unsafe.Sizeof(*v2))))
	} else if v1 != nil {
		block, err = me.createStructBlock(parent, *v1, fmt.Sprintf("%s (V1)", label), subcontracts.RelativeAddress64(index*uint64(unsafe.Sizeof(*v1))))
	} else {
		return fmt.Errorf("unknown subcache structure")
	}
	if err != nil {
		return err
	}
	if index == 0 {
		err := addLink(headerBlock, "SubCacheArrayOffset", block, "points to")
		if err != nil {
			return err
		}
		if me.addSizeLink {
			err = addLink(headerBlock, "SubCacheArrayCount", block, "gives size")
			if err != nil {
				return err
			}
		}
	}
	return addLink(block, "CacheVmOffset", subCache, "points to")
}
