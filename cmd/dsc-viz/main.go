package main

import (
	"fmt"
	"os"

	"github.com/LouisBrunner/dsc-viz/pkg/contracts"
	"github.com/LouisBrunner/dsc-viz/pkg/fetch"
	"github.com/sirupsen/logrus"
)

func main() {
	success, err := work()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dsc-viz: %s\n", err.Error())
		os.Exit(2)
	}
	if !success {
		os.Exit(1)
	}
}

func work() (bool, error) {
	params, err := getArgs()
	if err != nil {
		return false, fmt.Errorf("failed to parse arguments: %w", err)
	}

	logger := logrus.New()
	if os.Getenv("DEBUG") != "" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetReportCaller(true)
	}
	logger.SetLevel(params.loggingLevel)
	logger.SetOutput(os.Stderr)

	var fetcher contracts.Fetcher
	if params.fromArch != "" {
		fetcher, err = fetch.ScanForFileWithArchitecture(logger, params.fromArch)
		if err != nil {
			return false, err
		}
	} else if params.fromFile != "" {
		fetcher, err = fetch.FromFile(logger, params.fromFile)
		if err != nil {
			return false, err
		}
	} else if params.fromCurrentArch {
		fetcher, err = fetch.ScanForFileWithCurrentArchitecture(logger)
		if err != nil {
			return false, err
		}
	} else if params.fromMemory {
		fetcher, err = fetch.FromMemory(logger)
		if err != nil {
			return false, err
		}
	} else {
		return false, fmt.Errorf("no source specified")
	}

	fmt.Printf("fetcher: %v\n", fetcher)
	// TODO: more processing
	return true, nil
}
