package viz

import (
	"encoding/json"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
)

func OutputJSON(m contracts.MemoryBlock) (string, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
