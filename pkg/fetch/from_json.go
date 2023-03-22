package fetch

import (
	"encoding/json"
	"io"
	"os"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

func FromJSONFile(logger *logrus.Logger, filename string) (*contracts.MemoryBlock, error) {
	var content []byte
	var err error
	// FIXME: I really don't like that this is handled here...
	// we should be passing a io.Reader to this func and handling this logic in the main IMO
	if filename == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(filename)
	}
	if err != nil {
		return nil, err
	}
	return FromJSONText(logger, string(content))
}

func FromJSONText(logger *logrus.Logger, text string) (*contracts.MemoryBlock, error) {
	var mb contracts.MemoryBlock
	err := json.Unmarshal([]byte(text), &mb)
	if err != nil {
		return nil, err
	}
	return &mb, nil
}
