package parse

import (
	"fmt"

	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func (me *parser) parseImage(frame *blockFrame, img arrayElement) error {
	imgData, cast := img.Data.(subcontracts.DYLDCacheImageInfo)
	if !cast {
		return fmt.Errorf("invalid image info type: %T", img)
	}

	fmt.Printf("%+v\n", imgData)

	return nil
}
