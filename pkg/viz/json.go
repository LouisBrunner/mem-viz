package viz

import (
	"encoding/json"

	"github.com/LouisBrunner/mem-viz/pkg/contracts"
)

func (me *outputter) JSON(m contracts.MemoryBlock) error {
	encoder := json.NewEncoder(me.w)
	return encoder.Encode(m)
}
