package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const page = 0x1000

type parser struct {
	logger                *logrus.Logger
	slide                 uint64
	addSizeLink           bool
	deepSearch            bool
	thresholdsArrayTooBig uint64
	uniqueBlocks          map[category][]*contracts.MemoryBlock
	allBlocks             map[uintptr]*[]*contracts.MemoryBlock
	root                  *contracts.MemoryBlock
}

func Parse(logger *logrus.Logger, fetcher subcontracts.Fetcher) (*contracts.MemoryBlock, error) {
	mainHeader := fetcher.Header()
	slide, err := calculateSlide(fetcher, mainHeader)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Slide: %#016x", slide)

	parser := parser{
		logger: logger,
		slide:  slide,
		// TODO: should be dsc-viz flags
		addSizeLink:           false,
		deepSearch:            false,
		thresholdsArrayTooBig: 3000,
		uniqueBlocks:          make(map[category][]*contracts.MemoryBlock),
		allBlocks:             make(map[uintptr]*[]*contracts.MemoryBlock),
	}
	return parser.parse(fetcher)
}

func (me *parser) parse(fetcher subcontracts.Fetcher) (*contracts.MemoryBlock, error) {
	root := &contracts.MemoryBlock{
		Name:         "DSC",
		Address:      fetcher.BaseAddress(),
		ParentOffset: uint64(fetcher.BaseAddress()),
	}
	me.root = root

	me.logger.Debugf("Parsing main header")
	mainHeader := fetcher.Header()
	mainBlock, headerBlock, err := me.addCache(root, fetcher, "Main Header", subcontracts.ManualAddress(0))
	if err != nil {
		return nil, err
	}

	anchors := map[*contracts.MemoryBlock]struct{}{
		mainBlock: {},
	}

	subCacheEntriesFn := make([]func(*contracts.MemoryBlock) error, 0, len(fetcher.SubCaches()))
	subCacheSize := uint64(0)
	for i, cache := range fetcher.SubCaches() {
		me.logger.Debugf("Parsing subcache %d\n", i)

		me.clearNonGlobalCategories()

		v2, v1 := cache.SubCacheHeader()
		name := fmt.Sprintf("%d", i+1)
		if v2 != nil {
			name = commons.FromCString(v2.FileSuffix[:])
		}
		var subBlock, subHeaderBlock *contracts.MemoryBlock
		subBlock, subHeaderBlock, err = me.addCache(root, cache, fmt.Sprintf("Sub Cache %s", name), subcontracts.RelativeAddress64(cache.BaseAddress()))
		if err != nil {
			return nil, err
		}
		anchors[subBlock] = struct{}{}

		if v2 != nil {
			subCacheSize += uint64(unsafe.Sizeof(*v2))
		} else if v1 != nil {
			subCacheSize += uint64(unsafe.Sizeof(*v1))
		}

		subCacheEntriesFn = append(subCacheEntriesFn, func(i uint64) func(subCacheEntries *contracts.MemoryBlock) error {
			return func(subCacheEntries *contracts.MemoryBlock) error {
				return me.addSubCacheEntry(subCacheEntries, headerBlock, subHeaderBlock, fetcher.Header(), v2, v1, uint64(i))
			}
		}(uint64(i)))
	}

	if len(subCacheEntriesFn) > 0 {
		// A bit convoluted but it allows to have the size inside the block instead of an empty block (which is for more loose grouping)
		subCacheEntries, err := me.createCommonBlock(mainBlock, fmt.Sprintf("Subcache Entries (%d)", len(subCacheEntriesFn)), mainHeader.SubCacheArrayOffset, subCacheSize)
		if err != nil {
			return nil, err
		}
		for _, fn := range subCacheEntriesFn {
			err = fn(subCacheEntries)
			if err != nil {
				return nil, err
			}
		}
	}

	me.logger.Debugf("Rebalancing\n")
	me.flushCategories()
	me.rebalance(root, anchors)

	me.logger.Debugf("Parsing finished\n")

	return root, nil
}

func (me *parser) rebalance(root *contracts.MemoryBlock, anchors map[*contracts.MemoryBlock]struct{}) {
	addresses := maps.Keys(me.allBlocks)
	slices.Sort(addresses)

	if len(root.Content) < 1 {
		return
	}

	baseAnchor := 0
	for _, address := range addresses {
		anchor := baseAnchor
		for i := baseAnchor; i < len(root.Content); i += 1 {
			if root.Content[i].Address > address {
				break
			}
			anchor = i
		}
		baseAnchor = anchor

		parent := root.Content[anchor]
		sameAddress := me.allBlocks[address]
		for _, block := range *sameAddress {
			newParent := addChildDeep(parent, block)
			block.ParentOffset = uint64(block.Address - newParent.Address)
		}
	}
}

func calculateSlide(cache subcontracts.Cache, header subcontracts.DYLDCacheHeaderV3) (uint64, error) {
	reader := header.MappingOffset.GetReader(cache, 0, 0)
	mapping := &subcontracts.DYLDCacheMappingInfo{}
	err := commons.Unpack(reader, mapping)
	if err != nil {
		return 0, fmt.Errorf("failed to unpack mapping info to calculate slide: %w", err)
	}
	return uint64(cache.BaseAddress() - uintptr(mapping.Address)), nil
}
