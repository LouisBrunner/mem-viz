package parse

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
)

func Parse(fetcher contracts.Fetcher) (*contracts.MemoryBlock, error) {
	root := &contracts.MemoryBlock{
		Name:    "DSC",
		Address: fetcher.BaseAddress(),
	}
	mainHeader := fetcher.Header()
	mainCache := addHeader(root, mainHeader, "Main Header", 0)
	for i, cache := range fetcher.SubCaches() {
		v2, v1 := cache.SubCacheHeader()
		name := fmt.Sprintf("%d", i+1)
		if v2 != nil {
			name = strings.TrimRight(string(v2.FileSuffix[:]), "\x00")
		}
		subCache := addHeader(root, cache.Header(), fmt.Sprintf("Sub Cache %s", name), uint64(cache.BaseAddress()))
		err := addSubCache(mainCache, mainHeader, subCache, v2, v1, uint64(i))
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

func addHeader(parent *contracts.MemoryBlock, header contracts.DYLDCacheHeaderV3, label string, offset uint64) *contracts.MemoryBlock {
	var block *contracts.MemoryBlock
	if v1, ok := header.V1(); ok {
		block = createBlock(parent, v1, fmt.Sprintf("%s (V1)", label), offset)
	} else if v2, ok := header.V2(); ok {
		block = createBlock(parent, v2, fmt.Sprintf("%s (V2)", label), offset)
	} else {
		block = createBlock(parent, header, fmt.Sprintf("%s (V3)", label), offset)
	}
	return block
}

func addSubCache(parent *contracts.MemoryBlock, header contracts.DYLDCacheHeaderV3, subCache *contracts.MemoryBlock, v2 *contracts.DYLDSubcacheEntryV2, v1 *contracts.DYLDSubcacheEntryV1, index uint64) error {
	var block *contracts.MemoryBlock
	label := "Subcache Entry"
	if v2 != nil {
		block = createBlock(parent, *v2, fmt.Sprintf("%s (V2)", label), uint64(header.SubCacheArrayOffset)+index*uint64(unsafe.Sizeof(*v2)))
	} else if v1 != nil {
		block = createBlock(parent, *v1, fmt.Sprintf("%s (V1)", label), uint64(header.SubCacheArrayOffset)+index*uint64(unsafe.Sizeof(*v1)))
	} else {
		return fmt.Errorf("unknown subcache structure")
	}
	if index == 0 {
		err := addLink(parent, "SubCacheArrayOffset", block, "points to")
		if err != nil {
			return err
		}
		err = addLink(parent, "SubCacheArrayCount", block, "gives size")
		if err != nil {
			return err
		}
	}
	return addLink(block, "CacheVmOffset", subCache, "points to")
}
