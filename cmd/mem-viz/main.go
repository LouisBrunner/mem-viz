package main

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/cli"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/sirupsen/logrus"
)

func main() {
	cli.Main("mem-viz", nil, cli.Worker[interface{}]{
		GetMemory: func(logger *logrus.Logger, params interface{}) (*contracts.MemoryBlock, error) {
			return nil, fmt.Errorf("missing from flag: %s", cli.FromCommonSourcesHelp)
		},
	})
}
