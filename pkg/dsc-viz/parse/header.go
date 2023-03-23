package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func addCache(parent *contracts.MemoryBlock, cache subcontracts.Cache, label string, offset uint64) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	block := createEmptyBlock(parent, fmt.Sprintf("%s Area", label), offset)

	header := cache.Header()
	v1, okV1 := header.V1()
	v2, okV2 := header.V2()
	var headerBlock *contracts.MemoryBlock
	if okV1 {
		headerBlock = createStructBlock(block, v1, fmt.Sprintf("%s (V1)", label), 0)
	} else if okV2 {
		headerBlock = createStructBlock(block, v2, fmt.Sprintf("%s (V2)", label), 0)
	} else {
		headerBlock = createStructBlock(block, header, fmt.Sprintf("%s (V3)", label), 0)
	}
	_, err := parseAndAddMultipleStructs(cache, block, headerBlock, "MappingOffset", uint64(header.MappingOffset), "MappingCount", uint64(header.MappingCount), subcontracts.DYLDCacheMappingInfo{}, "Mappings")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each mapping
	if okV1 {
		_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "ImagesOffset", uint64(v1.ImagesOffset), "ImagesCount", uint64(v1.ImagesCount), subcontracts.DYLDCacheImageInfo{}, "Images")
		if err != nil {
			return nil, nil, err
		}
		// TODO: dig deeper in each image
	}
	_, err = createBlobBlock(block, headerBlock, "CodeSignatureOffset", uint64(header.CodeSignatureOffset), "CodeSignatureSize", uint64(header.CodeSignatureSize), "Code Signature")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should it use DYLDCacheSlideInfo and related?
		_, err = createBlobBlock(block, headerBlock, "SlideInfoOffset", uint64(v1.SlideInfoOffset), "SlideInfoSize", uint64(v1.SlideInfoSize), "Slide Info")
		if err != nil {
			return nil, nil, err
		}
	}
	// FIXME: should it use DYLDCacheLocalSymbolsInfo and related?
	_, err = createBlobBlock(block, headerBlock, "LocalSymbolsOffset", uint64(header.LocalSymbolsOffset), "LocalSymbolsSize", uint64(header.LocalSymbolsSize), "Local Symbols")
	if err != nil {
		return nil, nil, err
	}
	// FIXME: is it worth unpacking each uint64 and list them?
	_, err = createBlobBlock(block, headerBlock, "BranchPoolsOffset", uint64(header.BranchPoolsOffset), "BranchPoolsCount", uint64(header.BranchPoolsCount), "Branch Pools")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should it use DYLDCacheAcceleratorInfo and related?
		_, err = createBlobBlock(block, headerBlock, "AccelerateInfoAddr", uint64(v1.AccelerateInfoAddr), "AccelerateInfoSize", uint64(v1.AccelerateInfoSize), "Accelerate Info")
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = addLink(headerBlock, "DyldInCacheMh", &contracts.MemoryBlock{Address: block.Address + uintptr(header.DyldInCacheMh)}, "points to")
		if err != nil {
			return nil, nil, err
		}
		err = addLink(headerBlock, "DyldInCacheEntry", &contracts.MemoryBlock{Address: block.Address + uintptr(header.DyldInCacheEntry)}, "points to")
		if err != nil {
			return nil, nil, err
		}
	}
	_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "ImagesTextOffset", uint64(header.ImagesTextOffset), "ImagesTextCount", uint64(header.ImagesTextCount), subcontracts.DYLDCacheImageTextInfo{}, "Images Text")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each image text
	if okV1 {
		// FIXME: should it use a struct?
		_, err = createBlobBlock(block, headerBlock, "DylibsImageGroupAddr", uint64(v1.DylibsImageGroupAddr), "DylibsImageGroupSize", uint64(v1.DylibsImageGroupSize), "Dylibs ImageGroups")
		if err != nil {
			return nil, nil, err
		}
		_, err = createBlobBlock(block, headerBlock, "OtherImageGroupAddr", uint64(v1.OtherImageGroupAddr), "OtherImageGroupSize", uint64(v1.OtherImageGroupSize), "Other ImageGroups")
		if err != nil {
			return nil, nil, err
		}
	} else {
		_, err = createBlobBlock(block, headerBlock, "PatchInfoAddr", uint64(header.PatchInfoAddr), "PatchInfoSize", uint64(header.PatchInfoSize), "Patch Info")
		if err != nil {
			return nil, nil, err
		}
		// TODO: add dyld_cache_patch_info and related
	}
	_, err = createBlobBlock(block, headerBlock, "ProgClosuresAddr", uint64(header.ProgClosuresAddr), "ProgClosuresSize", uint64(header.ProgClosuresSize), "Program Closures")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgClosuresTrieAddr", uint64(header.ProgClosuresTrieAddr), "ProgClosuresTrieSize", uint64(header.ProgClosuresTrieSize), "Program Closures (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "DylibsImageArrayAddr", uint64(header.DylibsImageArrayAddr), "DylibsImageArraySize", uint64(header.DylibsImageArraySize), "Dylibs ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "DylibsTrieAddr", uint64(header.DylibsTrieAddr), "DylibsTrieSize", uint64(header.DylibsTrieSize), "Dylibs ImageArrays (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "OtherImageArrayAddr", uint64(header.OtherImageArrayAddr), "OtherImageArraySize", uint64(header.OtherImageArraySize), "Other ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "OtherTrieAddr", uint64(header.OtherTrieAddr), "OtherTrieSize", uint64(header.OtherTrieSize), "Other ImageArrays (Trie)")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		return block, headerBlock, nil
	}
	_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "MappingWithSlideOffset", uint64(header.MappingWithSlideOffset), "MappingWithSlideCount", uint64(header.MappingWithSlideCount), subcontracts.DYLDCacheMappingAndSlideInfo{}, "Mappings With Slide")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each mapping with slide
	err = addLink(headerBlock, "DylibsPblSetAddr", &contracts.MemoryBlock{Address: block.Address + uintptr(header.DylibsPblSetAddr)}, "points to")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgramsPblSetPoolAddr", uint64(header.ProgramsPblSetPoolAddr), "ProgramsPblSetPoolSize", uint64(header.ProgramsPblSetPoolSize), "PrebuiltLoaderSet for each program")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgramTrieAddr", uint64(header.ProgramTrieAddr), "ProgramTrieSize", uint64(header.ProgramTrieSize), "PrebuiltLoaderSet for each program (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "SwiftOptsOffset", uint64(header.SwiftOptsOffset), "SwiftOptsSize", uint64(header.SwiftOptsSize), "Swift Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "RosettaReadOnlyAddr", uint64(header.RosettaReadOnlyAddr), "RosettaReadOnlySize", uint64(header.RosettaReadOnlySize), "Rosetta Read-Only Region")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "RosettaReadWriteAddr", uint64(header.RosettaReadWriteAddr), "RosettaReadWriteSize", uint64(header.RosettaReadWriteSize), "Rosetta Read-Write Region")
	if err != nil {
		return nil, nil, err
	}
	_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "ImagesOffset", uint64(header.ImagesOffset), "ImagesCount", uint64(header.ImagesCount), subcontracts.DYLDCacheImageInfo{}, "Images")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each image
	if okV2 {
		return block, headerBlock, nil
	}
	_, err = createBlobBlock(block, headerBlock, "ObjcOptsOffset", uint64(header.ObjcOptsOffset), "ObjcOptsSize", uint64(header.ObjcOptsSize), "Objective-C Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "CacheAtlasOffset", uint64(header.CacheAtlasOffset), "CacheAtlasSize", uint64(header.CacheAtlasSize), "Cache Atlas")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "DynamicDataOffset", uint64(header.DynamicDataOffset), "DynamicDataMaxSize", uint64(header.DynamicDataMaxSize), "DYLD Cache Dynamic Data")
	if err != nil {
		return nil, nil, err
	}
	// FIXME: somehow we can't read this :(
	// _, err = parseAndAddStruct(cache, dcdd, headerBlock, "DynamicDataHeader", uint64(header.DynamicDataOffset), subcontracts.DYLDCacheDynamicDataHeader{}, "DYLD Cache Dynamic Data Header")
	// if err != nil {
	// 	return nil, nil, err
	// }
	return block, headerBlock, nil
}

func addSubCacheEntry(parent, headerBlock, subCache *contracts.MemoryBlock, header subcontracts.DYLDCacheHeaderV3, v2 *subcontracts.DYLDSubcacheEntryV2, v1 *subcontracts.DYLDSubcacheEntryV1, index uint64) error {
	var block *contracts.MemoryBlock
	label := "Subcache Entry"
	if v2 != nil {
		block = createStructBlock(parent, *v2, fmt.Sprintf("%s (V2)", label), index*uint64(unsafe.Sizeof(*v2)))
	} else if v1 != nil {
		block = createStructBlock(parent, *v1, fmt.Sprintf("%s (V1)", label), index*uint64(unsafe.Sizeof(*v1)))
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
