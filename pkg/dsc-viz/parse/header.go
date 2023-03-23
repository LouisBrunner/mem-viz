package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func addCache(parent *contracts.MemoryBlock, cache subcontracts.Cache, label string, offset uint64) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	block := createBlock[interface{}](parent, nil, fmt.Sprintf("%s Area", label), offset)

	header := cache.Header()
	var headerBlock *contracts.MemoryBlock
	if v1, ok := header.V1(); ok {
		headerBlock = createBlock(block, v1, fmt.Sprintf("%s (V1)", label), 0)
	} else if v2, ok := header.V2(); ok {
		headerBlock = createBlock(block, v2, fmt.Sprintf("%s (V2)", label), 0)
	} else {
		headerBlock = createBlock(block, header, fmt.Sprintf("%s (V3)", label), 0)
	}
	_, err := parseAndAddMultipleStructs(cache, block, headerBlock, "MappingOffset", uint64(header.MappingOffset), "MappingCount", uint64(header.MappingCount), subcontracts.DYLDCacheMappingInfo{}, "Mapping")
	if err != nil {
		return nil, nil, err
	}
	return block, headerBlock, nil
}

func addSubCacheEntry(parent, headerBlock, subCache *contracts.MemoryBlock, header subcontracts.DYLDCacheHeaderV3, v2 *subcontracts.DYLDSubcacheEntryV2, v1 *subcontracts.DYLDSubcacheEntryV1, index uint64) error {
	var block *contracts.MemoryBlock
	label := "Subcache Entry"
	if v2 != nil {
		block = createBlock(parent, *v2, fmt.Sprintf("%s (V2)", label), index*uint64(unsafe.Sizeof(*v2)))
	} else if v1 != nil {
		block = createBlock(parent, *v1, fmt.Sprintf("%s (V1)", label), index*uint64(unsafe.Sizeof(*v1)))
	} else {
		return fmt.Errorf("unknown subcache structure")
	}
	if index == 0 {
		err := addLink(headerBlock, "SubCacheArrayOffset", block, "points to")
		if err != nil {
			return err
		}
		err = addLink(headerBlock, "SubCacheArrayCount", block, "gives size")
		if err != nil {
			return err
		}
	}
	return addLink(block, "CacheVmOffset", subCache, "points to")
}
