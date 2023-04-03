package parse

import (
	"fmt"
	"unsafe"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseMappingsCommon(frame *blockFrame, mappingHeaders []arrayElement, prevMappings []*contracts.MemoryBlock, apply func(frame *blockFrame, img arrayElement, prevMapping *contracts.MemoryBlock) (*contracts.MemoryBlock, error)) ([]*contracts.MemoryBlock, error) {
	hasAllMappings := len(mappingHeaders) == len(prevMappings)
	if len(prevMappings) != 0 && !hasAllMappings {
		return nil, fmt.Errorf("mismatched number of mappings: %d != %d", len(mappingHeaders), len(prevMappings))
	}
	var err error
	for i, mapping := range mappingHeaders {
		var prevMapping *contracts.MemoryBlock
		if hasAllMappings {
			prevMapping = prevMappings[i]
		}
		prevMapping, err = apply(frame, mapping, prevMapping)
		if err != nil {
			return nil, err
		}
		if !hasAllMappings {
			prevMappings = append(prevMappings, prevMapping)
		}
	}
	return prevMappings, nil
}

func (me *parser) parseMappings(frame *blockFrame, mappingHeaders []arrayElement, prevMappings []*contracts.MemoryBlock) ([]*contracts.MemoryBlock, error) {
	return me.parseMappingsCommon(frame, mappingHeaders, prevMappings, me.parseMapping)
}

func (me *parser) parseMapping(frame *blockFrame, mapping arrayElement, prevMapping *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	mappingData, cast := mapping.Data.(subcontracts.DYLDCacheMappingInfo)
	if !cast {
		return nil, fmt.Errorf("invalid mapping info type: %T", mapping.Data)
	}

	return me.parseMappingCommon(frame, mapping.Block, commonMappingData{
		Address:    mappingData.Address,
		Size:       mappingData.Size,
		FileOffset: mappingData.FileOffset,
	}, prevMapping)
}

func (me *parser) parseMappingsWithSlide(frame *blockFrame, mappingHeaders []arrayElement, prevMappings []*contracts.MemoryBlock) ([]*contracts.MemoryBlock, error) {
	leMapping := mappingHeaders[len(mappingHeaders)-1]
	linkEdit, cast := leMapping.Data.(subcontracts.DYLDCacheMappingAndSlideInfo)
	if !cast {
		return nil, fmt.Errorf("invalid mapping with slide info type: %T", leMapping.Data)
	}

	return me.parseMappingsCommon(frame, mappingHeaders, prevMappings, func(frame *blockFrame, mapping arrayElement, prevMapping *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
		return me.parseMappingWithSlide(frame, mapping, prevMapping, linkEdit)
	})
}

func (me *parser) parseMappingWithSlide(frame *blockFrame, mapping arrayElement, prevMapping *contracts.MemoryBlock, linkEdit subcontracts.DYLDCacheMappingAndSlideInfo) (*contracts.MemoryBlock, error) {
	mappingData, cast := mapping.Data.(subcontracts.DYLDCacheMappingAndSlideInfo)
	if !cast {
		return nil, fmt.Errorf("invalid mapping with slide info type: %T", mapping.Data)
	}

	prevMapping, err := me.parseMappingCommon(frame, mapping.Block, commonMappingData{
		Address:    mappingData.Address,
		Size:       mappingData.Size,
		FileOffset: mappingData.FileOffset,
	}, prevMapping)
	if err != nil {
		return nil, err
	}

	if mappingData.SlideInfoFileOffset == 0 {
		return prevMapping, nil
	}

	offset := mappingData.SlideInfoFileOffset - linkEdit.FileOffset
	newAddress := linkEdit.Address + subcontracts.UnslidAddress(offset)
	reader := newAddress.GetReader(frame.cache, 0, me.slide)
	versionHeader := subcontracts.DYLDCacheSlideInfoVersion{}
	err = commons.Unpack(reader, &versionHeader)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("%s.Slide Info (V%d)", mapping.Block.Name, versionHeader.Version)
	short := uint16(0)
	shortPtr := &short

	sideFrame := frame.siblingFrame(mapping.Block)
	switch versionHeader.Version {
	case 1:
		slideInfoV1 := &subcontracts.DYLDCacheSlideInfo{}
		blob, header, err := me.parseAndAddBlob(sideFrame, "SlideInfoFileOffset", newAddress, "SlideInfoFileSize", mappingData.SlideInfoFileSize, slideInfoV1, title)
		if err != nil {
			return nil, err
		}
		subFrame := frame.pushFrame(blob, header)
		_, _, err = me.parseAndAddArray(subFrame, "TocOffset", newAddress+subcontracts.UnslidAddress(slideInfoV1.TocOffset), "TocCount", uint64(slideInfoV1.TocCount), shortPtr, fmt.Sprintf("%s.TOC", title))
		if err != nil {
			return nil, err
		}
		_, _, err = me.parseAndAddArray(subFrame, "EntriesOffset", newAddress+subcontracts.UnslidAddress(slideInfoV1.EntriesOffset), "EntriesCount", uint64(slideInfoV1.EntriesCount), subcontracts.DYLDCacheSlideInfoEntry{}, fmt.Sprintf("%s.Entries", title))
		if err != nil {
			return nil, err
		}
	case 2:
		slideInfoV2 := &subcontracts.DYLDCacheSlideInfo2{}
		blob, header, err := me.parseAndAddBlob(sideFrame, "SlideInfoFileOffset", newAddress, "SlideInfoFileSize", mappingData.SlideInfoFileSize, slideInfoV2, title)
		if err != nil {
			return nil, err
		}
		subFrame := frame.pushFrame(blob, header)
		_, _, err = me.parseAndAddArray(subFrame, "PageStartsOffset", newAddress+subcontracts.UnslidAddress(slideInfoV2.PageStartsOffset), "PageStartsCount", uint64(slideInfoV2.PageStartsCount), shortPtr, fmt.Sprintf("%s.Pages", title))
		if err != nil {
			return nil, err
		}
		_, _, err = me.parseAndAddArray(subFrame, "PageExtrasOffset", newAddress+subcontracts.UnslidAddress(slideInfoV2.PageExtrasOffset), "PageExtrasCount", uint64(slideInfoV2.PageExtrasCount), shortPtr, fmt.Sprintf("%s.Extra Pages", title))
		if err != nil {
			return nil, err
		}
	case 3:
		slideInfoV3 := &subcontracts.DYLDCacheSlideInfo3{}
		blob, header, err := me.parseAndAddBlob(sideFrame, "SlideInfoFileOffset", newAddress, "SlideInfoFileSize", mappingData.SlideInfoFileSize, slideInfoV3, title)
		if err != nil {
			return nil, err
		}
		subFrame := frame.pushFrame(blob, header)
		_, _, err = me.parseAndAddArray(subFrame, "", newAddress+subcontracts.UnslidAddress(unsafe.Sizeof(*slideInfoV3)), "PageStartsCount", uint64(slideInfoV3.PageStartsCount), shortPtr, fmt.Sprintf("%s.Pages", title))
		if err != nil {
			return nil, err
		}
	case 4:
		slideInfoV4 := &subcontracts.DYLDCacheSlideInfo4{}
		blob, header, err := me.parseAndAddBlob(sideFrame, "SlideInfoFileOffset", newAddress, "SlideInfoFileSize", mappingData.SlideInfoFileSize, slideInfoV4, title)
		if err != nil {
			return nil, err
		}
		subFrame := frame.pushFrame(blob, header)
		_, _, err = me.parseAndAddArray(subFrame, "PageStartsOffset", newAddress+subcontracts.UnslidAddress(slideInfoV4.PageStartsOffset), "PageStartsCount", uint64(slideInfoV4.PageStartsCount), shortPtr, fmt.Sprintf("%s.Pages", title))
		if err != nil {
			return nil, err
		}
		_, _, err = me.parseAndAddArray(subFrame, "PageExtrasOffset", newAddress+subcontracts.UnslidAddress(slideInfoV4.PageExtrasOffset), "PageExtrasCount", uint64(slideInfoV4.PageExtrasCount), shortPtr, fmt.Sprintf("%s.Extra Pages", title))
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported slide info version: %d", versionHeader.Version)
	}
	return prevMapping, nil
}

type commonMappingData struct {
	Address    subcontracts.UnslidAddress
	Size       uint64
	FileOffset subcontracts.RelativeAddress64
}

func (me *parser) parseMappingCommon(frame *blockFrame, mapping *contracts.MemoryBlock, data commonMappingData, prevMapping *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	sideFrame := frame.siblingFrame(mapping)

	err := me.addLinkWithOffset(sideFrame, "Address", data.Address, "points to")
	if err != nil {
		return nil, err
	}
	if me.addSizeLink {
		err = me.addLinkWithOffset(sideFrame, "Size", data.Address, "gives size")
		if err != nil {
			return nil, err
		}
	}

	err = me.addLinkWithOffset(sideFrame, "FileOffset", data.FileOffset, "points to")
	if err != nil {
		return nil, err
	}

	if prevMapping == nil {
		prevMapping, err = me.createCommonBlock(sideFrame.parent, mapping.Name, data.Address, data.Size)
		if err != nil {
			return nil, err
		}
	}

	return prevMapping, nil
}
