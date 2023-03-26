package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

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

	// FIXME: we could dig deeper in those structs but all we would get is a bunch of extra links (with some actual strings sometimes)
	// but there are so many entries and it's so specific to the patch table that I don't see the point

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
	case 2:
		blob, patchHeaderBlock, err = me.parseAndAddBlob(frame, "PatchInfoAddr", header.PatchInfoAddr, "PatchInfoSize", header.PatchInfoSize, &patchInfoV2, "Patch Info (V2)")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown patch table version: %d", patchHeader.PatchTableVersion)
	}

	frame = frame.pushFrame(blob, patchHeaderBlock)

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
	_, _, err = me.parseAndAddArray(frame, "PatchLocationArrayAddr", patchInfoV2.PatchLocationArrayAddr, "PatchLocationArrayCount", uint64(patchInfoV2.PatchLocationArrayCount), &subcontracts.DYLDCachePatchableLocationV2{}, "Patch Locations")
	if err != nil {
		return err
	}
	_, err = me.createBlobBlock(frame, "PatchExportNamesAddr", patchInfoV2.PatchExportNamesAddr, "PatchExportNamesSize", uint64(patchInfoV2.PatchExportNamesSize), "Patch Export Names")
	if err != nil {
		return err
	}
	return nil
}
