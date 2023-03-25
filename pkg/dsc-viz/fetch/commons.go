package fetch

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
)

func checkMagic(header contracts.DYLDCacheHeaderV3) error {
	if string(header.Magic[:])[0:7] != contracts.MAGIC_arm64[0:7] {
		return fmt.Errorf("invalid magic %s", commons.FromCString(header.Magic[:]))
	}
	return nil
}
