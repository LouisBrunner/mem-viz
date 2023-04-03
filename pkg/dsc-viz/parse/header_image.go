package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseImagesCommon(frame *blockFrame, imgs []arrayElement, pathBlock *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock, apply func(frame *blockFrame, img arrayElement, pathBlock *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error)) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	var err error
	for _, img := range imgs {
		pathBlock, imgBlocks, err = apply(frame, img, pathBlock, imgBlocks)
		if err != nil {
			return nil, nil, err
		}
	}
	return pathBlock, imgBlocks, nil
}

func (me *parser) parseImages(frame *blockFrame, imgs []arrayElement, pathBlock *contracts.MemoryBlock, imgBlock []*contracts.MemoryBlock) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	return me.parseImagesCommon(frame, imgs, pathBlock, imgBlock, me.parseImage)
}

func (me *parser) parseImage(frame *blockFrame, img arrayElement, pathBlock *contracts.MemoryBlock, imgBlock []*contracts.MemoryBlock) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageInfo)
	if !cast {
		return nil, nil, fmt.Errorf("invalid image info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, pathBlock, imgBlock, "Address", imgData.Address, "", 0, "PathFileOffset", imgData.PathFileOffset)
}

func (me *parser) parseImageTexts(frame *blockFrame, imgs []arrayElement, pathBlock *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	return me.parseImagesCommon(frame, imgs, pathBlock, imgBlocks, me.parseImageText)
}

func (me *parser) parseImageText(frame *blockFrame, img arrayElement, pathBlock *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageTextInfo)
	if !cast {
		return nil, nil, fmt.Errorf("invalid image text info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, pathBlock, imgBlocks, "LoadAddress", imgData.LoadAddress, "Size", uint64(imgData.TextSegmentSize), "PathOffset", imgData.PathOffset)
}

func (me *parser) addImage(frame *blockFrame, image, pathBlock *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock, addressName string, address subcontracts.Address, sizeName string, size uint64, pathOffsetName string, pathOffset subcontracts.Address) (*contracts.MemoryBlock, []*contracts.MemoryBlock, error) {
	path := readCString(pathOffset.GetReader(frame.cache, 0, me.slide))
	var err error

	// FIXME: not sure about this, quite noisy also abusing the 0/0
	// addValue(img.Block, "Path", path, 0, 0)

	imgBlocks, err = me.addImageTEXT(frame, image, imgBlocks, path, addressName, address, sizeName, size)
	if err != nil {
		return nil, nil, err
	}

	pathBlock, err = me.addImagePath(frame, image, pathBlock, path, pathOffsetName, pathOffset)
	if err != nil {
		return nil, nil, err
	}

	return pathBlock, imgBlocks, nil
}

func (me *parser) addImageTEXT(frame *blockFrame, image *contracts.MemoryBlock, imgBlocks []*contracts.MemoryBlock, path, addressName string, address subcontracts.Address, sizeName string, size uint64) ([]*contracts.MemoryBlock, error) {
	var err error
	const page = 0x1000

	absAddr := address.AddBase(frame.parent.Address).Calculate(me.slide)
	var currBlock *contracts.MemoryBlock
	for _, imgBlock := range imgBlocks {
		if isInsideOf(&contracts.MemoryBlock{Address: absAddr}, &contracts.MemoryBlock{Address: imgBlock.Address, Size: roundUp(imgBlock.GetSize(), page)}) {
			currBlock = imgBlock
			break
		}
	}
	if currBlock == nil {
		currBlock, err = me.createEmptyBlock(frame.parent, "Images TEXT", address)
		if err != nil {
			return nil, err
		}
		imgBlocks = append(imgBlocks, currBlock)
	}

	// TODO: should we add a struct (MH?)
	if size == 0 {
		sideFrame := frame.siblingFrame(image)
		err = me.addLinkWithOffset(sideFrame, addressName, address, "points to")
		if err != nil {
			return nil, err
		}
		return imgBlocks, nil
	}

	imgFrame := frame.pushFrame(currBlock, image)
	frameBasedAddress := subcontracts.ManualAddress(absAddr - currBlock.Address)
	for _, imgB := range currBlock.Content {
		if imgB.Address == absAddr {
			err = me.addLinkWithOffset(imgFrame, addressName, frameBasedAddress, "points to")
			if err != nil {
				return nil, err
			}
			return imgBlocks, nil
		} else if imgB.Address > absAddr {
			break
		}
	}

	_, err = me.createBlobBlock(imgFrame, addressName, address, sizeName, size, fmt.Sprintf("%s TEXT", path))
	if err != nil {
		return nil, err
	}
	currBlock.Name = fmt.Sprintf("Images TEXT (%d)", len(currBlock.Content))
	currBlock.Size = 0
	currBlock.Size = currBlock.GetSize()
	return imgBlocks, nil
}

func (me *parser) addImagePath(frame *blockFrame, image, pathBlock *contracts.MemoryBlock, path, pathOffsetName string, pathOffset subcontracts.Address) (*contracts.MemoryBlock, error) {
	var err error

	absAddr := pathOffset.AddBase(frame.parent.Address).Calculate(me.slide)
	if pathBlock == nil || !isInsideOf(&contracts.MemoryBlock{Address: absAddr}, pathBlock) {
		pathBlock, err = me.createEmptyBlock(frame.parent, "Paths", pathOffset)
		if err != nil {
			return nil, err
		}
	}

	pathFrame := frame.pushFrame(pathBlock, image)
	frameBasedAddress := subcontracts.ManualAddress(absAddr - pathBlock.Address)
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
	pathBlock.Size = 0
	pathBlock.Size = pathBlock.GetSize()
	return pathBlock, nil
}
