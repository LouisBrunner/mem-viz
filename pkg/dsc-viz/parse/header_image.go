package parse

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/parsingutils"
)

func (me *parser) parseImagesCommon(frame *blockFrame, imgs []arrayElement, apply func(frame *blockFrame, img arrayElement) error) error {
	for _, img := range imgs {
		err := apply(frame, img)
		if err != nil {
			return err
		}
	}
	return nil
}

func (me *parser) parseImages(frame *blockFrame, imgs []arrayElement) error {
	return me.parseImagesCommon(frame, imgs, me.parseImage)
}

func (me *parser) parseImage(frame *blockFrame, img arrayElement) error {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageInfo)
	if !cast {
		return fmt.Errorf("invalid image info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, "Address", imgData.Address, "", 0, "PathFileOffset", imgData.PathFileOffset)
}

func (me *parser) parseImageTexts(frame *blockFrame, imgs []arrayElement) error {
	return me.parseImagesCommon(frame, imgs, me.parseImageText)
}

func (me *parser) parseImageText(frame *blockFrame, img arrayElement) error {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageTextInfo)
	if !cast {
		return fmt.Errorf("invalid image text info type: %T", img.Data)
	}

	return me.addImage(frame, img.Block, "LoadAddress", imgData.LoadAddress, "Size", uint64(imgData.TextSegmentSize), "PathOffset", imgData.PathOffset)
}

func (me *parser) addImage(frame *blockFrame, image *contracts.MemoryBlock, addressName string, address subcontracts.Address, sizeName string, size uint64, pathOffsetName string, pathOffset subcontracts.Address) error {
	path := readCString(pathOffset.GetReader(frame.cache, 0, me.slide))

	// FIXME: not sure about this, quite noisy also abusing the 0/0
	// addValue(img.Block, "Path", path, 0, 0)

	err := me.addImageTEXT(frame, image, path, addressName, address, sizeName, size)
	if err != nil {
		return err
	}

	err = me.addImagePath(frame, image, path, pathOffsetName, pathOffset)
	if err != nil {
		return err
	}

	return nil
}

func (me *parser) addImageTEXT(frame *blockFrame, image *contracts.MemoryBlock, path, addressName string, address subcontracts.Address, sizeName string, size uint64) error {
	absAddr := address.AddBase(frame.parent.Address).Calculate(me.slide)
	currBlock, err := me.findOrCreateUniqueBlock(categoryImages, func(i int, block *contracts.MemoryBlock) bool {
		return parsingutils.IsInsideOf(&contracts.MemoryBlock{Address: absAddr}, &contracts.MemoryBlock{Address: block.Address, Size: roundUp(block.GetSize(), page)})
	}, func() (*contracts.MemoryBlock, error) {
		return me.createEmptyBlock(frame.parent, "Images TEXT", address)
	})
	if err != nil {
		return err
	}

	if size == 0 {
		sideFrame := frame.siblingFrame(image)
		err := me.addLinkWithOffset(sideFrame, addressName, address, "points to")
		if err != nil {
			return err
		}
		return nil
	}

	imgFrame := frame.pushFrame(currBlock, image)
	frameBasedAddress := subcontracts.ManualAddress(absAddr - currBlock.Address)
	for _, imgB := range currBlock.Content {
		if imgB.Address == absAddr {
			err := me.addLinkWithOffset(imgFrame, addressName, frameBasedAddress, "points to")
			if err != nil {
				return err
			}
			return nil
		} else if imgB.Address > absAddr {
			break
		}
	}

	text, err := me.createBlobBlock(imgFrame, addressName, address, sizeName, size, fmt.Sprintf("%s TEXT", path))
	if err != nil {
		return err
	}
	_, err = me.parseMachO(imgFrame, text, path)
	if err != nil {
		return err
	}

	me.updateCategoryBlock(currBlock, "Images TEXT")
	return nil
}

func (me *parser) addImagePath(frame *blockFrame, image *contracts.MemoryBlock, path, pathOffsetName string, pathOffset subcontracts.Address) error {
	absAddr := pathOffset.AddBase(frame.parent.Address).Calculate(me.slide)
	pathBlock, err := me.findOrCreateUniqueBlock(categoryPaths, func(i int, pathBlock *contracts.MemoryBlock) bool {
		return parsingutils.IsInsideOf(&contracts.MemoryBlock{Address: absAddr}, pathBlock)
	}, func() (*contracts.MemoryBlock, error) {
		return me.createEmptyBlock(frame.parent, "Paths", pathOffset)
	})
	if err != nil {
		return err
	}

	pathFrame := frame.pushFrame(pathBlock, image)
	frameBasedAddress := subcontracts.ManualAddress(absAddr - pathBlock.Address)
	for _, pathB := range pathBlock.Content {
		if pathB.Address == absAddr {
			err := me.addLinkWithOffset(pathFrame, pathOffsetName, frameBasedAddress, "points to")
			if err != nil {
				return err
			}
			return nil
		} else if pathB.Address > absAddr {
			break
		}
	}

	_, err = me.createBlobBlock(pathFrame, pathOffsetName, frameBasedAddress, "", uint64(len(path)+1), fmt.Sprintf("Path: %s", path))
	if err != nil {
		return err
	}

	me.updateCategoryBlock(pathBlock, "Paths")
	return nil
}
