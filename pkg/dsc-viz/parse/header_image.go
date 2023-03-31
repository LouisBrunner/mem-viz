package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseImagesCommon(frame *blockFrame, imgs []arrayElement, pathsBlock *contracts.MemoryBlock, apply func(frame *blockFrame, img arrayElement, pathsBlock *contracts.MemoryBlock) (*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, error) {
	var err error
	for _, img := range imgs {
		pathsBlock, err = apply(frame, img, pathsBlock)
		if err != nil {
			return nil, err
		}
	}
	return pathsBlock, nil
}

func (me *parser) parseImages(frame *blockFrame, imgs []arrayElement, pathBlock *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	return me.parseImagesCommon(frame, imgs, pathBlock, me.parseImage)
}

func (me *parser) parseImage(frame *blockFrame, img arrayElement, pathBlock *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageInfo)
	if !cast {
		return nil, fmt.Errorf("invalid image info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, pathBlock, "Address", imgData.Address, "", 0, "PathFileOffset", imgData.PathFileOffset)
}

func (me *parser) parseImageTexts(frame *blockFrame, imgs []arrayElement, pathBlocks *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	return me.parseImagesCommon(frame, imgs, pathBlocks, me.parseImageText)
}

func (me *parser) parseImageText(frame *blockFrame, img arrayElement, pathBlock *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageTextInfo)
	if !cast {
		return nil, fmt.Errorf("invalid image text info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, pathBlock, "LoadAddress", imgData.LoadAddress, "Size", uint64(imgData.TextSegmentSize), "PathOffset", imgData.PathOffset)
}

func (me *parser) addImage(frame *blockFrame, image *contracts.MemoryBlock, pathBlock *contracts.MemoryBlock, addressName string, address subcontracts.Address, sizeName string, size uint64, pathOffsetName string, pathOffset subcontracts.Address) (*contracts.MemoryBlock, error) {
	path := readCString(pathOffset.GetReader(frame.cache, 0, me.slide))
	var err error

	// FIXME: not sure about this, quite noisy also abusing the 0/0
	// addValue(img.Block, "Path", path, 0, 0)

	// TODO: should we add a struct (MH?)
	sideFrame := frame.siblingFrame(image)
	if size == 0 {
		err = me.addLinkWithOffset(sideFrame, addressName, address, "points to")
	} else {
		_, err = me.createBlobBlock(sideFrame, addressName, address, sizeName, size, fmt.Sprintf("%s TEXT", path))
	}
	if err != nil {
		return nil, err
	}

	absAddr := pathOffset.AddBase(frame.parent.Address).Calculate(me.slide)
	if pathBlock == nil || !(pathBlock.Address <= absAddr && absAddr <= pathBlock.Address+uintptr(pathBlock.GetSize())) {
		pathBlock, err = me.createEmptyBlock(frame.parent, "Paths", pathOffset)
		if err != nil {
			return nil, err
		}
	}

	pathFrame := frame.pushFrame(pathBlock, image)
	frameBasedAddress := subcontracts.ManualAddress(uintptr(pathOffset.AddBase(frame.parent.Address).Calculate(me.slide)) - pathBlock.Address)
	for _, pathB := range pathBlock.Content {
		if pathB.Address == absAddr {
			err = me.addLinkWithOffset(pathFrame, pathOffsetName, frameBasedAddress, "points to")
			if err != nil {
				return nil, err
			}
			return pathBlock, nil
		} else if pathB.Address > absAddr {
			break
		}
	}

	_, err = me.createBlobBlock(pathFrame, pathOffsetName, frameBasedAddress, "", uint64(len(path)+1), fmt.Sprintf("Path: %s", path))
	if err != nil {
		return nil, err
	}
	pathBlock.Name = fmt.Sprintf("Paths (%d)", len(pathBlock.Content))
	return pathBlock, nil
}
