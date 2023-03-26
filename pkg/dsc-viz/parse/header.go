package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func addCache(parent *contracts.MemoryBlock, cache subcontracts.Cache, label string, offset uint64) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	block, err := createEmptyBlock(parent, fmt.Sprintf("%s Area", label), offset)
	if err != nil {
		return nil, nil, err
	}

	header := cache.Header()
	v1, okV1 := header.V1()
	v2, okV2 := header.V2()
	var headerBlock *contracts.MemoryBlock
	if okV1 {
		headerBlock, err = createStructBlock(block, v1, fmt.Sprintf("%s (V1)", label), uint64(0))
	} else if okV2 {
		headerBlock, err = createStructBlock(block, v2, fmt.Sprintf("%s (V2)", label), uint64(0))
	} else {
		headerBlock, err = createStructBlock(block, header, fmt.Sprintf("%s (V3)", label), uint64(0))
	}
	if err != nil {
		return nil, nil, err
	}

	_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "MappingOffset", uint64(header.MappingOffset), "MappingCount", uint64(header.MappingCount), subcontracts.DYLDCacheMappingInfo{}, "Mappings")
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
	_, err = createBlobBlock(block, headerBlock, "CodeSignatureOffset", header.CodeSignatureOffset, "CodeSignatureSize", header.CodeSignatureSize, "Code Signature")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should use DYLDCacheSlideInfo1,2,3 but don't have a V1 cache
		// https://github.com/apple-oss-distributions/dyld/blob/c8a445f88f9fc1713db34674e79b00e30723e79d/dyld/SharedCacheRuntime.cpp#L654
		_, err = createBlobBlock(block, headerBlock, "SlideInfoOffset", uint64(v1.SlideInfoOffset), "SlideInfoSize", uint64(v1.SlideInfoSize), "Slide Info")
		if err != nil {
			return nil, nil, err
		}
	}
	// FIXME: should use DYLDCacheLocalSymbolsInfo but my cache doesn't have them
	// https://github.com/apple-oss-distributions/dyld/blob/c8a445f88f9fc1713db34674e79b00e30723e79d/cache-builder/OptimizerLinkedit.cpp#L137
	_, err = createBlobBlock(block, headerBlock, "LocalSymbolsOffset", header.LocalSymbolsOffset, "LocalSymbolsSize", header.LocalSymbolsSize, "Local Symbols")
	if err != nil {
		return nil, nil, err
	}
	// FIXME: is it worth unpacking each uint64 and list them? probably too noisy
	_, err = createBlobBlock(block, headerBlock, "BranchPoolsOffset", uint64(header.BranchPoolsOffset), "BranchPoolsCount", uint64(header.BranchPoolsCount), "Branch Pools")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should it use DYLDCacheAcceleratorInfo? can't find reference to it in DYLD and don't have a V1 cache
		_, err = createBlobBlock(block, headerBlock, "AccelerateInfoAddr", uint64(v1.AccelerateInfoAddr), "AccelerateInfoSize", uint64(v1.AccelerateInfoSize), "Accelerate Info")
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = addLinkWithOffset(headerBlock, "DyldInCacheMh", header.DyldInCacheMh, "points to")
		if err != nil {
			return nil, nil, err
		}
		err = addLinkWithOffset(headerBlock, "DyldInCacheEntry", header.DyldInCacheEntry, "points to")
		if err != nil {
			return nil, nil, err
		}
	}
	_, err = parseAndAddMultipleStructs(cache, block, headerBlock, "ImagesTextOffset", header.ImagesTextOffset, "ImagesTextCount", header.ImagesTextCount, subcontracts.DYLDCacheImageTextInfo{}, "Images Text")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each image text
	if okV1 {
		// FIXME: should it use a struct? can't find reference to it in DYLD and don't have a V1 cache
		_, err = createBlobBlock(block, headerBlock, "DylibsImageGroupAddr", uint64(v1.DylibsImageGroupAddr), "DylibsImageGroupSize", uint64(v1.DylibsImageGroupSize), "Dylibs ImageGroups")
		if err != nil {
			return nil, nil, err
		}
		_, err = createBlobBlock(block, headerBlock, "OtherImageGroupAddr", uint64(v1.OtherImageGroupAddr), "OtherImageGroupSize", uint64(v1.OtherImageGroupSize), "Other ImageGroups")
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = parsePatchInfo(cache, block, headerBlock, header)
		if err != nil {
			return nil, nil, err
		}
	}
	_, err = createBlobBlock(block, headerBlock, "ProgClosuresAddr", header.ProgClosuresAddr, "ProgClosuresSize", header.ProgClosuresSize, "Program Closures")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgClosuresTrieAddr", header.ProgClosuresTrieAddr, "ProgClosuresTrieSize", header.ProgClosuresTrieSize, "Program Closures (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "DylibsImageArrayAddr", header.DylibsImageArrayAddr, "DylibsImageArraySize", header.DylibsImageArraySize, "Dylibs ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "DylibsTrieAddr", header.DylibsTrieAddr, "DylibsTrieSize", header.DylibsTrieSize, "Dylibs ImageArrays (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "OtherImageArrayAddr", header.OtherImageArrayAddr, "OtherImageArraySize", header.OtherImageArraySize, "Other ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "OtherTrieAddr", header.OtherTrieAddr, "OtherTrieSize", header.OtherTrieSize, "Other ImageArrays (Trie)")
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
	err = addLinkWithOffset(headerBlock, "DylibsPblSetAddr", header.DylibsPblSetAddr, "points to")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgramsPblSetPoolAddr", header.ProgramsPblSetPoolAddr, "ProgramsPblSetPoolSize", header.ProgramsPblSetPoolSize, "PrebuiltLoaderSet for each program")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "ProgramTrieAddr", header.ProgramTrieAddr, "ProgramTrieSize", uint64(header.ProgramTrieSize), "PrebuiltLoaderSet for each program (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "SwiftOptsOffset", header.SwiftOptsOffset, "SwiftOptsSize", header.SwiftOptsSize, "Swift Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "RosettaReadOnlyAddr", header.RosettaReadOnlyAddr, "RosettaReadOnlySize", header.RosettaReadOnlySize, "Rosetta Read-Only Region")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "RosettaReadWriteAddr", header.RosettaReadWriteAddr, "RosettaReadWriteSize", header.RosettaReadWriteSize, "Rosetta Read-Write Region")
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
	_, err = createBlobBlock(block, headerBlock, "ObjcOptsOffset", header.ObjcOptsOffset, "ObjcOptsSize", header.ObjcOptsSize, "Objective-C Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = createBlobBlock(block, headerBlock, "CacheAtlasOffset", header.CacheAtlasOffset, "CacheAtlasSize", header.CacheAtlasSize, "Cache Atlas")
	if err != nil {
		return nil, nil, err
	}
	blob, subHeaderBlock, subHeader, err := parseAndAddBlob(cache, block, headerBlock, "DynamicDataOffset", header.DynamicDataOffset, "DynamicDataMaxSize", header.DynamicDataMaxSize, subcontracts.DYLDCacheDynamicDataHeader{}, "DYLD Cache Dynamic Data")
	if err != nil {
		return nil, nil, err
	}
	if subHeaderBlock != nil && string(subHeader.Magic[:]) != subcontracts.DYLD_SHARED_CACHE_DYNAMIC_DATA_MAGIC {
		err = findAndRemoveChild(blob, subHeaderBlock)
		if err != nil {
			return nil, nil, err
		}
	}
	return block, headerBlock, nil
}

func addSubCacheEntry(parent, headerBlock, subCache *contracts.MemoryBlock, header subcontracts.DYLDCacheHeaderV3, v2 *subcontracts.DYLDSubcacheEntryV2, v1 *subcontracts.DYLDSubcacheEntryV1, index uint64) error {
	var block *contracts.MemoryBlock
	var err error
	label := "Subcache Entry"
	if v2 != nil {
		block, err = createStructBlock(parent, *v2, fmt.Sprintf("%s (V2)", label), index*uint64(unsafe.Sizeof(*v2)))
	} else if v1 != nil {
		block, err = createStructBlock(parent, *v1, fmt.Sprintf("%s (V1)", label), index*uint64(unsafe.Sizeof(*v1)))
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
		err = addLink(headerBlock, "SubCacheArrayCount", block, "gives size")
		if err != nil {
			return err
		}
	}
	return addLink(block, "CacheVmOffset", subCache, "points to")
}

func parsePatchInfo(cache subcontracts.Cache, block, headerBlock *contracts.MemoryBlock, header subcontracts.DYLDCacheHeaderV3) error {
	if header.PatchInfoAddr == 0 {
		return nil
	}

	if _, v1 := header.V1(); v1 {
		_, _, _, err := parseAndAddBlob(cache, block, headerBlock, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, subcontracts.DYLDCachePatchInfoV1{}, "Patch Info (V1)")
		if err != nil {
			return err
		}
		// FIXME: should add all related structs but don't have a V1 cache
		return nil
	}

	fmt.Printf("%#16x\n", header.PatchInfoAddr)
	reader := getReaderAtOffset(cache, header.PatchInfoAddr, 0)
	patchHeader := subcontracts.DYLDCachePatchInfo{}
	err := commons.Unpack(reader, &patchHeader)
	if err != nil {
		return err
	}

	if patchHeader.PatchTableVersion == 3 {
		_, _, _, err = parseAndAddBlob(cache, block, headerBlock, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, subcontracts.DYLDCachePatchInfoV3{}, "Patch Info (V3)")
		if err != nil {
			return err
		}
		// TODO: add dyld_cache_patch_info and related
	} else if patchHeader.PatchTableVersion == 2 {
		_, _, _, err = parseAndAddBlob(cache, block, headerBlock, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, subcontracts.DYLDCachePatchInfoV2{}, "Patch Info (V2)")
		if err != nil {
			return err
		}
		// TODO: add dyld_cache_patch_info and related
	} else {
		return fmt.Errorf("unknown patch table version: %d", patchHeader.PatchTableVersion)
	}
	return nil
}
