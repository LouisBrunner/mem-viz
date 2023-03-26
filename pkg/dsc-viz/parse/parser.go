package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/sirupsen/logrus"
)

var gSlide uint64

func Parse(logger *logrus.Logger, fetcher subcontracts.Fetcher) (*contracts.MemoryBlock, error) {
	root := &contracts.MemoryBlock{
		Name:    "DSC",
		Address: fetcher.BaseAddress(),
	}

	mainHeader := fetcher.Header()
	slide, err := calculateSlide(fetcher, mainHeader)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Slide: %#16x", slide)

	// TODO: HORRID!!!!!!!!!!!! (but it works)
	// find a way to pass the slide to the other functions without adding it on EVERY FUNCTION
	// context.Context? could add "current parent", "current header" etc
	gSlide = slide

	mainBlock, headerBlock, err := addCache(root, fetcher, "Main Header", 0)
	if err != nil {
		return nil, err
	}

	var subCacheEntries *contracts.MemoryBlock
	if l := len(fetcher.SubCaches()); l > 0 {
		subCacheEntries, err = createEmptyBlock(mainBlock, fmt.Sprintf("Subcache Entries (%d)", l), uint64(mainHeader.SubCacheArrayOffset))
		if err != nil {
			return nil, err
		}
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

	rebalance(root)

	return root, nil
}

func rebalance(root *contracts.MemoryBlock) {
	commons.VisitEachBlock(root, func(depth int, block *contracts.MemoryBlock) error {
		for i := 0; i < len(block.Content); i += 1 {
			child := block.Content[i]
			if i >= len(block.Content)-1 {
				break
			}
			previousChild := block.Content[i+1]
			for i := 0; i < len(child.Content); {
				gchild := child.Content[i]
				if gchild.Address+uintptr(gchild.GetSize()) > previousChild.Address {
					moveChild(child, block, i)
					continue
				}
				i += 1
			}
		}
		return nil
	})
}

func calculateSlide(cache subcontracts.Cache, header subcontracts.DYLDCacheHeaderV3) (uint64, error) {
	reader := getReaderAtOffset(cache, uint64(header.MappingOffset), 0)
	mapping := &subcontracts.DYLDCacheMappingInfo{}
	err := commons.Unpack(reader, mapping)
	if err != nil {
		return 0, nil
	}
	return uint64(cache.BaseAddress() - uintptr(mapping.Address)), nil
}
