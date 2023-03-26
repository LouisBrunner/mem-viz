package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
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
		headerBlock, err = me.createStructBlock(block, v1, fmt.Sprintf("%s (V1)", label), subcontracts.RelativeAddress32(0))
	} else if okV2 {
		headerBlock, err = me.createStructBlock(block, v2, fmt.Sprintf("%s (V2)", label), subcontracts.RelativeAddress32(0))
	} else {
		headerBlock, err = me.createStructBlock(block, header, fmt.Sprintf("%s (V3)", label), subcontracts.RelativeAddress32(0))
	}
	if err != nil {
		return nil, nil, err
	}

	frame := topFrame(cache, block, headerBlock)
	_, _, err = me.parseAndAddArray(frame, "MappingOffset", header.MappingOffset, "MappingCount", uint64(header.MappingCount), &subcontracts.DYLDCacheMappingInfo{}, "Mappings")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each mapping
	if okV1 {
		_, _, err = me.parseAndAddArray(frame, "ImagesOffset", v1.ImagesOffset, "ImagesCount", uint64(v1.ImagesCount), &subcontracts.DYLDCacheImageInfo{}, "Images")
		if err != nil {
			return nil, nil, err
		}
		// TODO: dig deeper in each image
	}
	_, err = me.createBlobBlock(frame, "CodeSignatureOffset", header.CodeSignatureOffset, "CodeSignatureSize", header.CodeSignatureSize, "Code Signature")
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
		err = me.addLinkWithOffset(headerBlock, "DyldInCacheMh", header.DyldInCacheMh, "points to")
		if err != nil {
			return nil, nil, err
		}
		err = me.addLinkWithOffset(headerBlock, "DyldInCacheEntry", header.DyldInCacheEntry, "points to")
		if err != nil {
			return nil, nil, err
		}
	}
	_, _, err = me.parseAndAddArray(frame, "ImagesTextOffset", header.ImagesTextOffset, "ImagesTextCount", header.ImagesTextCount, &subcontracts.DYLDCacheImageTextInfo{}, "Images Text")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each image text
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
	_, _, err = me.parseAndAddArray(frame, "MappingWithSlideOffset", header.MappingWithSlideOffset, "MappingWithSlideCount", uint64(header.MappingWithSlideCount), &subcontracts.DYLDCacheMappingAndSlideInfo{}, "Mappings With Slide")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each mapping with slide
	err = me.addLinkWithOffset(headerBlock, "DylibsPblSetAddr", header.DylibsPblSetAddr, "points to")
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
	_, _, err = me.parseAndAddArray(frame, "ImagesOffset", header.ImagesOffset, "ImagesCount", uint64(header.ImagesCount), &subcontracts.DYLDCacheImageInfo{}, "Images")
	if err != nil {
		return nil, nil, err
	}
	// TODO: dig deeper in each image
	if okV2 {
		return block, headerBlock, nil
	}
	_, err = me.createBlobBlock(frame, "ObjcOptsOffset", header.ObjcOptsOffset, "ObjcOptsSize", header.ObjcOptsSize, "Objective-C Optimizations Header")
	if err != nil {
		return nil, nil, err
	}
	_, err = me.createBlobBlock(frame, "CacheAtlasOffset", header.CacheAtlasOffset, "CacheAtlasSize", header.CacheAtlasSize, "Cache Atlas")
	if err != nil {
		return nil, nil, err
	}
	dcdd := subcontracts.DYLDCacheDynamicDataHeader{}
	dcddBlock, dcddHeaderBlock, err := me.parseAndAddBlob(frame, "DynamicDataOffset", header.DynamicDataOffset, "DynamicDataMaxSize", header.DynamicDataMaxSize, &dcdd, "DYLD Cache Dynamic Data")
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
		err = addLink(headerBlock, "SubCacheArrayCount", block, "gives size")
		if err != nil {
			return err
		}
	}
	return addLink(block, "CacheVmOffset", subCache, "points to")
}

func (me *parser) parsePatchInfo(frame *blockFrame, header subcontracts.DYLDCacheHeaderV3) error {
	if header.PatchInfoAddr == 0 {
		return nil
	}

	if _, v1 := header.V1(); v1 {
		_, _, err := me.parseAndAddBlob(frame, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, &subcontracts.DYLDCachePatchInfoV1{}, "Patch Info (V1)")
		if err != nil {
			return err
		}
		// FIXME: should add all related structs but don't have a V1 cache
		return nil
	}

	reader := header.PatchInfoAddr.GetReader(frame.cache, 0, me.slide)
	patchHeader := subcontracts.DYLDCachePatchInfo{}
	err := commons.Unpack(reader, &patchHeader)
	if err != nil {
		return err
	}

	var patchInfoV2 subcontracts.DYLDCachePatchInfoV2
	var blob, patchHeaderBlock *contracts.MemoryBlock

	switch patchHeader.PatchTableVersion {
	case 3:
		patchHeaderV3 := &subcontracts.DYLDCachePatchInfoV3{}
		blob, patchHeaderBlock, err = me.parseAndAddBlob(frame, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, patchHeaderV3, "Patch Info (V3)")
		if err != nil {
			return err
		}
		patchInfoV2 = patchHeaderV3.DYLDCachePatchInfoV2

		frame = frame.pushFrame(blob, patchHeaderBlock)
		_, _, err = me.parseAndAddArray(frame, "GotClientsArrayAddr", patchHeaderV3.GotClientsArrayAddr, "GotClientsArrayCount", uint64(patchHeaderV3.GotClientsArrayCount), &subcontracts.DYLDCacheImageGotClientsV3{}, "GOT Clients")
		if err != nil {
			return err
		}
		_, _, err = me.parseAndAddArray(frame, "GotClientExportsArrayAddr", patchHeaderV3.GotClientExportsArrayAddr, "GotClientExportsArrayCount", uint64(patchHeaderV3.GotClientExportsArrayCount), &subcontracts.DYLDCachePatchableExportV3{}, "GOT Client Exports")
		if err != nil {
			return err
		}
		_, _, err = me.parseAndAddArray(frame, "GotLocationArrayAddr", patchHeaderV3.GotLocationArrayAddr, "GotLocationArrayCount", uint64(patchHeaderV3.GotLocationArrayCount), &subcontracts.DYLDCachePatchableLocationV3{}, "GOT Locations")
		if err != nil {
			return err
		}

		// TODO: add dyld_cache_patch_info and related
	case 2:
		blob, patchHeaderBlock, err = me.parseAndAddBlob(frame, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, &patchInfoV2, "Patch Info (V2)")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown patch table version: %d", patchHeader.PatchTableVersion)
	}

	frame = frame.pushFrame(blob, patchHeaderBlock)

	// FIXME: too noisy to add all the structs
	// TODO: dig deeper still?
	_, _, err = me.parseAndAddArray(frame, "PatchTableArrayAddr", patchInfoV2.PatchTableArrayAddr, "PatchTableArrayCount", uint64(patchInfoV2.PatchTableArrayCount), &subcontracts.DYLDCacheImagePatchesV2{}, "Patch Table")
	if err != nil {
		return err
	}
	_, _, err = me.parseAndAddArray(frame, "PatchImageExportsArrayAddr", patchInfoV2.PatchImageExportsArrayAddr, "PatchImageExportsArrayCount", uint64(patchInfoV2.PatchImageExportsArrayCount), &subcontracts.DYLDCacheImageExportV2{}, "Patch Image Exports")
	if err != nil {
		return err
	}
	_, _, err = me.parseAndAddArray(frame, "PatchClientsArrayAddr", patchInfoV2.PatchClientsArrayAddr, "PatchClientsArrayCount", uint64(patchInfoV2.PatchClientsArrayCount), &subcontracts.DYLDCacheImageClientsV2{}, "Patch Clients")
	if err != nil {
		return err
	}
	_, _, err = me.parseAndAddArray(frame, "PatchClientExportsArrayAddr", patchInfoV2.PatchClientExportsArrayAddr, "PatchClientExportsArrayCount", uint64(patchInfoV2.PatchClientExportsArrayCount), &subcontracts.DYLDCachePatchableExportV2{}, "Patch Client Exports")
	if err != nil {
		return err
	}
	// _, _, err = me.parseAndAddArray(frame, "PatchLocationArrayAddr", patchInfoV2.PatchLocationArrayAddr, "PatchLocationArrayCount", uint64(patchInfoV2.PatchLocationArrayCount), &subcontracts.DYLDCachePatchableLocationV2{}, "Patch Locations")
	// if err != nil {
	// 	return err
	// }
	return nil
}
