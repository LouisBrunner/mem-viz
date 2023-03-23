package parse

import (
	"fmt"

	"github.com/LouisBrunner/dsc-viz/pkg/commons"
	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
)

func Parse(fetcher contracts.Fetcher) (*contracts.MemoryBlock, error) {
	root := &contracts.MemoryBlock{
		Name:    "DSC",
		Address: fetcher.BaseAddress(),
	}
	mainHeader := fetcher.Header()

	mainBlock, headerBlock, err := addCache(root, fetcher, "Main Header", 0)
	if err != nil {
		return nil, err
	}

	var subCacheEntries *contracts.MemoryBlock
	if l := len(fetcher.SubCaches()); l > 0 {
		subCacheEntries = createBlock[interface{}](mainBlock, nil, fmt.Sprintf("Subcache Entries (%d)", l), uint64(mainHeader.SubCacheArrayOffset))
	}
	for i, cache := range fetcher.SubCaches() {
		v2, v1 := cache.SubCacheHeader()
		name := fmt.Sprintf("%d", i+1)
		if v2 != nil {
			name = commons.FromCString(v2.FileSuffix[:])
		}
		_, subHeaderBlock, err := addCache(root, cache, fmt.Sprintf("Sub Cache %s", name), uint64(cache.BaseAddress()))
		if err != nil {
			return nil, err
		}
		err = addSubCacheEntry(subCacheEntries, headerBlock, subHeaderBlock, fetcher.Header(), v2, v1, uint64(i))
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}
