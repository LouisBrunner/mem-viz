package viz

import (
	"encoding/json"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
)

func (me *outputter) JSON(m contracts.MemoryBlock) error {
	raw, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = me.w.Write(raw)
	return err
}
