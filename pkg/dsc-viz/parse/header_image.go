package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseImages(frame *blockFrame, imgs []arrayElement) error {
	var pathsBlock *contracts.MemoryBlock
	var err error

	for _, img := range imgs {
		pathsBlock, err = me.parseImage(frame, img, pathsBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (me *parser) parseImage(frame *blockFrame, img arrayElement, previousPathBlock *contracts.MemoryBlock) (*contracts.MemoryBlock, error) {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageInfo)
	if !cast {
		return nil, fmt.Errorf("invalid image info type: %T", img)
	}

	pathOffset := imgData.PathFileOffset
	path := readCString(pathOffset.GetReader(frame.cache, 0, me.slide))

	// FIXME: not sure about this, quite noisy also abusing the 0/0
	// addValue(img.Block, "Path", path, 0, 0)

	// TODO: should we add a struct (MH?)
	sideFrame := frame.siblingFrame(img.Block)
	err := me.addLinkWithOffset(sideFrame, "Address", imgData.Address, "points to")
	if err != nil {
		return nil, err
	}

	pathBlock := previousPathBlock
	if pathBlock == nil || pathBlock.Address+uintptr(pathBlock.GetSize()) != pathOffset.AddBase(frame.parent.Address).Calculate(me.slide) {
		pathBlock, err = me.createEmptyBlock(frame.parent, "Paths", pathOffset)
		if err != nil {
			return nil, err
		}
	}

	pathFrame := frame.pushFrame(pathBlock, img.Block)
	frameBasedAddress := subcontracts.ManualAddress(uintptr(pathOffset.AddBase(frame.parent.Address).Calculate(me.slide)) - pathBlock.Address)
	_, err = me.createBlobBlock(pathFrame, "PathFileOffset", frameBasedAddress, "", uint64(len(path)+1), fmt.Sprintf("Path: %s", path))
	if err != nil {
		return nil, err
	}

	pathBlock.Name = fmt.Sprintf("Paths (%d)", len(pathBlock.Content))

	return pathBlock, nil
}
