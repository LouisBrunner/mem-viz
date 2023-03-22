package main

import (
	"fmt"
	"os"

	"github.com/LouisBrunner/dsc-viz/pkg/fetch"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	if os.Getenv("DEBUG") != "" {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetReportCaller(true)
	}
	logger.SetOutput(os.Stderr)

	mem, err := fetch.FromMemory(logger)
	fmt.Printf("mem: %v, err: %v\n", mem, err)
	carch, err := fetch.ScanForFileWithCurrentArchitecture(logger)
	fmt.Printf("carch: %v, err: %v\n", carch, err)
	arch, err := fetch.ScanForFileWithArchitecture(logger, "arm64")
	fmt.Printf("arch: %v, err: %v\n", arch, err)
	file, err := fetch.FromFile(logger, "/System/Cryptexes/OS/System/Library/dyld/dyld_shared_cache_arm64e")
	fmt.Printf("file: %v, err: %v\n", file, err)
}
