package main

import (
	"fmt"

	"github.com/LouisBrunner/mem-viz/pkg/cli"
	"github.com/LouisBrunner/mem-viz/pkg/contracts"
	subcontracts "github.com/LouisBrunner/mem-viz/pkg/dsc-viz/contracts"
	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/fetch"
	"github.com/LouisBrunner/mem-viz/pkg/dsc-viz/parse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type args struct {
	fromArch        string
	fromFile        string
	fromCurrentArch bool
	fromMemory      bool
}

func main() {
	cli.Main("dsc-viz", args{}, cli.Worker[args]{
		AddFlags: func(params *args) {
			pflag.StringVar(&params.fromArch, "from-arch", "", "architecture of the file to load")
			pflag.StringVar(&params.fromFile, "from-file", "", "file to load")
			pflag.BoolVar(&params.fromCurrentArch, "from-current-arch", false, "load the file for the current architecture")
			pflag.BoolVar(&params.fromMemory, "from-memory", false, "load the memory from the current process")
		},
		CheckExtraFrom: func(params args) ([]bool, []string) {
			return []bool{
					params.fromArch != "",
					params.fromFile != "",
					params.fromCurrentArch,
					params.fromMemory,
				}, []string{
					"from-arch",
					"from-file",
					"from-current-arch",
					"from-memory",
				}
		},
		GetMemory: func(logger *logrus.Logger, params args) (*contracts.MemoryBlock, error) {
			var fetcher subcontracts.Fetcher
			var err error

			if params.fromArch != "" {
				fetcher, err = fetch.ScanForFileWithArchitecture(logger, params.fromArch)
			} else if params.fromFile != "" {
				fetcher, err = fetch.FromFile(logger, params.fromFile)
			} else if params.fromCurrentArch {
				fetcher, err = fetch.ScanForFileWithCurrentArchitecture(logger)
			} else if params.fromMemory {
				fetcher, err = fetch.FromMemory(logger)
			} else {
				err = fmt.Errorf("no source specified")
			}
			if err != nil {
				return nil, err
			}

			return parse.Parse(fetcher)
		},
	})
}
