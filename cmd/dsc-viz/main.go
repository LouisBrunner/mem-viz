package main

import (
	"fmt"
	"os"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/LouisBrunner/dsc-viz/pkg/fetch"
	"github.com/LouisBrunner/dsc-viz/pkg/parse"
	"github.com/LouisBrunner/dsc-viz/pkg/viz"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	err := work()
	if err != nil {
		if err == pflag.ErrHelp {
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stderr, "dsc-viz: %s\n", err.Error())
			os.Exit(1)
		}
	}
}

func getFetcher(logger *logrus.Logger, params args) (contracts.Fetcher, *contracts.MemoryBlock, error) {
	var fetcher contracts.Fetcher
	var mb *contracts.MemoryBlock
	var err error

	if params.fromArch != "" {
		fetcher, err = fetch.ScanForFileWithArchitecture(logger, params.fromArch)
	} else if params.fromFile != "" {
		fetcher, err = fetch.FromFile(logger, params.fromFile)
	} else if params.fromCurrentArch {
		fetcher, err = fetch.ScanForFileWithCurrentArchitecture(logger)
	} else if params.fromMemory {
		fetcher, err = fetch.FromMemory(logger)
	} else if params.fromJSONFile != "" {
		mb, err = fetch.FromJSONFile(logger, params.fromJSONFile)
	} else if params.fromJSONText != "" {
		mb, err = fetch.FromJSONText(logger, params.fromJSONText)
	} else {
		err = fmt.Errorf("no source specified")
	}
	if err != nil {
		return nil, nil, err
	}

	return fetcher, mb, nil
}

func getOutput(logger *logrus.Logger, outputFormat, outputFile string) (func(mb contracts.MemoryBlock) error, func(), error) {
	var outputFn func(mb contracts.MemoryBlock) error
	cleanupFn := func() {}
	var err error

	w := os.Stdout
	if outputFile != "" {
		w, err = os.Create(outputFile)
		if err != nil {
			return nil, nil, err
		}
		cleanupFn = func() {
			w.Close()
		}
	}

	outputter := viz.New(logger, w)

	switch outputFormat {
	case outputFormatGraphviz:
		outputFn = outputter.Graphviz
	case outputFormatLaTeX:
		outputFn = outputter.LaTeX
	case outputFormatMarkdown:
		outputFn = outputter.Markdown
	case outputFormatText:
		outputFn = outputter.Text
	case outputFormatASCII:
		outputFn = outputter.ASCII
	case outputFormatJSON:
		outputFn = outputter.JSON
	default:
		err = fmt.Errorf("unknown output format: %s", outputFormat)
	}
	if err != nil {
		cleanupFn()
		return nil, nil, err
	}
	return outputFn, cleanupFn, nil
}

func work() error {
	params, err := getArgs()
	if err != nil {
		return err
	}

	logger := logrus.New()
	logger.SetLevel(params.loggingLevel)
	logger.SetOutput(os.Stderr)
	if os.Getenv("DEBUG") != "" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetReportCaller(true)
	}

	// FIXME: not loving this, would be better if it could fit in one interface
	fetcher, mb, err := getFetcher(logger, *params)
	if err != nil {
		return err
	}

	// If we loaded the MemoryBlocks directly from JSON, then no need to parse anything
	if mb == nil {
		mb, err = parse.Parse(fetcher)
		if err != nil {
			return err
		}
	} else {
		logger.Debug("skipping the parsing, we got JSON")
	}

	// TODO: check guarantees

	outputFn, cleanupFn, err := getOutput(logger, params.outputFormat, params.outputFile)
	if err != nil {
		return err
	}
	defer cleanupFn()

	return outputFn(*mb)
}
