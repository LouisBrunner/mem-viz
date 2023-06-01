package main

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/cli"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/macho-viz"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type args struct {
	file string
}

func main() {
	cli.Main("macho-viz", args{}, cli.Worker[args]{
		AddFlags: func(params *args) {
			pflag.StringVar(&params.file, "file", "", "file to load")
		},
		CheckExtraFrom: func(params args) ([]bool, []string) {
			return []bool{
					params.file != "",
				}, []string{
					"file",
				}
		},
		GetMemory: func(logger *logrus.Logger, params args) (*contracts.MemoryBlock, error) {
			if params.file == "" {
				return nil, fmt.Errorf("no source specified")
			}
			return macho.Parse(logger, params.file)
		},
	})
}
