package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) addCache(parent *contracts.MemoryBlock, cache subcontracts.Cache, label string, offset subcontracts.Address) (*contracts.MemoryBlock, *contracts.MemoryBlock, error) {
	block, err := me.createEmptyBlock(parent, fmt.Sprintf("%s Area", label), offset)
	if err != nil {
		return nil, nil, err
	}

	header := cache.Header()
	v1, okV1 := header.V1()
	v2, okV2 := header.V2()
	var headerBlock *contracts.MemoryBlock
	if okV1 {
		headerBlock, err = me.createStructBlock(block, v1, fmt.Sprintf("%s (V1)", label), subcontracts.ManualAddress(0))
	} else if okV2 {
		headerBlock, err = me.createStructBlock(block, v2, fmt.Sprintf("%s (V2)", label), subcontracts.ManualAddress(0))
	} else {
		headerBlock, err = me.createStructBlock(block, header, fmt.Sprintf("%s (V3)", label), subcontracts.ManualAddress(0))
	}
	if err != nil {
		return nil, nil, err
	}

	frame := topFrame(cache, block, headerBlock)
	var pathsBlock *contracts.MemoryBlock
	var mappingBlocks []*contracts.MemoryBlock
	_, mappings, err := me.parseAndAddArray(frame, "MappingOffset", header.MappingOffset, "MappingCount", uint64(header.MappingCount), &subcontracts.DYLDCacheMappingInfo{}, "Mappings")
	if err != nil {
		return nil, nil, err
	}
	mappingBlocks, err = me.parseMappings(frame, mappings, mappingBlocks)
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		_, imgs, err := me.parseAndAddArray(frame, "ImagesOffset", v1.ImagesOffset, "ImagesCount", uint64(v1.ImagesCount), &subcontracts.DYLDCacheImageInfo{}, "Images")
		if err != nil {
			return nil, nil, err
		}
		pathsBlock, err = me.parseImages(frame, imgs, pathsBlock)
		if err != nil {
			return nil, nil, err
		}
	}
	cs, err := me.createBlobBlock(frame, "CodeSignatureOffset", header.CodeSignatureOffset, "CodeSignatureSize", header.CodeSignatureSize, "Code Signature")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should use DYLDCacheSlideInfo1,2,3 but don't have a V1 cache
		// https://github.com/apple-oss-distributions/dyld/blob/c8a445f88f9fc1713db34674e79b00e30723e79d/dyld/SharedCacheRuntime.cpp#L654
		_, err = me.createBlobBlock(frame, "SlideInfoOffset", v1.SlideInfoOffset, "SlideInfoSize", uint64(v1.SlideInfoSize), "Slide Info")
		if err != nil {
			return nil, nil, err
		}
	}
	// FIXME: should use DYLDCacheLocalSymbolsInfo but my cache doesn't have them
	// https://github.com/apple-oss-distributions/dyld/blob/c8a445f88f9fc1713db34674e79b00e30723e79d/cache-builder/OptimizerLinkedit.cpp#L137
	_, err = me.createBlobBlock(frame, "LocalSymbolsOffset", header.LocalSymbolsOffset, "LocalSymbolsSize", header.LocalSymbolsSize, "Local Symbols")
	if err != nil {
		return nil, nil, err
	}
	// FIXME: is it worth unpacking each uint64 and list them? probably too noisy
	_, _, err = me.parseAndAddArray(frame, "BranchPoolsOffset", header.BranchPoolsOffset, "BranchPoolsCount", uint64(header.BranchPoolsCount), uint64(1), "Branch Pools")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should it use DYLDCacheAcceleratorInfo? can't find reference to it in DYLD and don't have a V1 cache
		_, err = me.createBlobBlock(frame, "AccelerateInfoAddr", v1.AccelerateInfoAddr, "AccelerateInfoSize", uint64(v1.AccelerateInfoSize), "Accelerate Info")
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = me.addLinkWithOffset(frame, "DyldInCacheMh", header.DyldInCacheMh, "points to")
		if err != nil {
			return nil, nil, err
		}
		err = me.addLinkWithOffset(frame, "DyldInCacheEntry", header.DyldInCacheEntry, "points to")
		if err != nil {
			return nil, nil, err
		}
	}
	_, imgTexts, err := me.parseAndAddArray(frame, "ImagesTextOffset", header.ImagesTextOffset, "ImagesTextCount", header.ImagesTextCount, &subcontracts.DYLDCacheImageTextInfo{}, "Images Text")
	if err != nil {
		return nil, nil, err
	}
	pathsBlock, err = me.parseImageTexts(frame, imgTexts, pathsBlock)
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		// FIXME: should it use a struct? can't find reference to it in DYLD and don't have a V1 cache
		_, err = me.createBlobBlock(frame, "DylibsImageGroupAddr", v1.DylibsImageGroupAddr, "DylibsImageGroupSize", uint64(v1.DylibsImageGroupSize), "Dylibs ImageGroups")
		if err != nil {
			return nil, nil, err
		}
		_, err = me.createBlobBlock(frame, "OtherImageGroupAddr", v1.OtherImageGroupAddr, "OtherImageGroupSize", uint64(v1.OtherImageGroupSize), "Other ImageGroups")
		if err != nil {
			return nil, nil, err
		}
	} else {
		err = me.parsePatchInfo(frame, header)
		if err != nil {
			return nil, nil, err
		}
	}
	_, err = me.createBlobBlock(frame, "ProgClosuresAddr", header.ProgClosuresAddr, "ProgClosuresSize", header.ProgClosuresSize, "Program Closures")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "ProgClosuresTrieAddr", header.ProgClosuresTrieAddr, "ProgClosuresTrieSize", header.ProgClosuresTrieSize, "Program Closures (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "DylibsImageArrayAddr", header.DylibsImageArrayAddr, "DylibsImageArraySize", header.DylibsImageArraySize, "Dylibs ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "DylibsTrieAddr", header.DylibsTrieAddr, "DylibsTrieSize", header.DylibsTrieSize, "Dylibs ImageArrays (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "OtherImageArrayAddr", header.OtherImageArrayAddr, "OtherImageArraySize", header.OtherImageArraySize, "Other ImageArrays")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "OtherTrieAddr", header.OtherTrieAddr, "OtherTrieSize", header.OtherTrieSize, "Other ImageArrays (Trie)")
	if err != nil {
		return nil, nil, err
	}
	if okV1 {
		return block, headerBlock, nil
	}
	_, mappingsWithSlide, err := me.parseAndAddArray(frame, "MappingWithSlideOffset", header.MappingWithSlideOffset, "MappingWithSlideCount", uint64(header.MappingWithSlideCount), &subcontracts.DYLDCacheMappingAndSlideInfo{}, "Mappings With Slide")
	if err != nil {
		return nil, nil, err
	}
	mappingBlocks, err = me.parseMappingsWithSlide(frame, mappingsWithSlide, mappingBlocks)
	if err != nil {
		return nil, nil, err
	}
	err = me.addLinkWithOffset(frame, "DylibsPblSetAddr", header.DylibsPblSetAddr, "points to")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "ProgramsPblSetPoolAddr", header.ProgramsPblSetPoolAddr, "ProgramsPblSetPoolSize", header.ProgramsPblSetPoolSize, "PrebuiltLoaderSet for each program")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "ProgramTrieAddr", header.ProgramTrieAddr, "ProgramTrieSize", uint64(header.ProgramTrieSize), "PrebuiltLoaderSet for each program (Trie)")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "SwiftOptsOffset", header.SwiftOptsOffset, "SwiftOptsSize", header.SwiftOptsSize, "Swift Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "RosettaReadOnlyAddr", header.RosettaReadOnlyAddr, "RosettaReadOnlySize", header.RosettaReadOnlySize, "Rosetta Read-Only Region")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "RosettaReadWriteAddr", header.RosettaReadWriteAddr, "RosettaReadWriteSize", header.RosettaReadWriteSize, "Rosetta Read-Write Region")
	if err != nil {
		return nil, nil, err
	}
	_, imgs, err := me.parseAndAddArray(frame, "ImagesOffset", header.ImagesOffset, "ImagesCount", uint64(header.ImagesCount), &subcontracts.DYLDCacheImageInfo{}, "Images")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.parseImages(frame, imgs, pathsBlock)
	if err != nil {
		return nil, nil, err
	}
	if okV2 {
		return block, headerBlock, nil
	}
	objcOptsSize := header.ObjcOptsSize
	if cs != nil {
		objcOptsAddr := header.ObjcOptsOffset.AddBase(frame.parent.Address).Calculate(me.slide)
		if cs.Address > objcOptsAddr && objcOptsAddr+uintptr(objcOptsSize) > cs.Address { // FIXME: amd64 has overlapping CS and ObjcOpts, somehow?
			objcOptsSize = uint64(cs.Address - objcOptsAddr)
		}
	}
	_, err = me.createBlobBlock(frame, "ObjcOptsOffset", header.ObjcOptsOffset, "ObjcOptsSize", objcOptsSize, "Objective-C Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "CacheAtlasOffset", header.CacheAtlasOffset, "CacheAtlasSize", header.CacheAtlasSize, "Cache Atlas")
	if err != nil {
		return nil, nil, err
	}
	dcdd := subcontracts.DYLDCacheDynamicDataHeader{}
	dcddSize := header.DynamicDataMaxSize
	if dcddSize > 0x400000 { // FIXME: amd64 has some super large values (Petabytes), no idea why and DYLD mostly seems to ignore the value?
		dcddSize = 0x1000
	}
	dcddBlock, dcddHeaderBlock, err := me.parseAndAddBlob(frame, "DynamicDataOffset", header.DynamicDataOffset, "DynamicDataMaxSize", dcddSize, &dcdd, "DYLD Cache Dynamic Data")
	if err != nil {
		return nil, nil, err
	}
	if dcddHeaderBlock != nil && string(dcdd.Magic[:]) != subcontracts.DYLD_SHARED_CACHE_DYNAMIC_DATA_MAGIC {
		err = findAndRemoveChild(dcddBlock, dcddHeaderBlock)
		if err != nil {
			return nil, nil, err
		}
	}
	return block, headerBlock, nil
}
