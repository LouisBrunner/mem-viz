package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

type args struct {
	fromArch        string
	fromCurrentArch bool
	fromFile        string
	fromMemory      bool
	loggingLevel    logrus.Level
}

func getArgs() (*args, error) {
	params := &args{
		loggingLevel: logrus.ErrorLevel,
	}

	pflag.StringVar(&params.fromArch, "from-arch", "", "architecture to fetch from (e.g. arm64)")
	pflag.BoolVar(&params.fromCurrentArch, "from-current-arch", false, "fetch from the current architecture")
	pflag.StringVar(&params.fromFile, "from-file", "", "file to fetch from")
	pflag.BoolVar(&params.fromMemory, "from-memory", false, "fetch from memory")
	loggingLevelStr := ""
	pflag.StringVar(&loggingLevelStr, "logging-level", "error", "logrus log level for internal debugging")
	pflag.Parse()

	if pflag.NArg() > 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(pflag.Args(), ", "))
	}

	// Parsing
	logLevel, err := logrus.ParseLevel(loggingLevelStr)
	if err != nil {
		return nil, err
	}
	params.loggingLevel = logLevel

	// Check that only one of the --from-* flags is set
	fromFlagsCount := map[bool]int{}
	for _, flag := range []bool{
		params.fromArch != "",
		params.fromCurrentArch,
		params.fromFile != "",
		params.fromMemory,
	} {
		fromFlagsCount[flag] += 1
	}
	if fromFlagsCount[true] > 1 {
		return nil, fmt.Errorf("cannot specify more than one of --from-arch, --from-current-arch, --from-file, --from-memory")
	}
	if fromFlagsCount[true] == 0 {
		return nil, fmt.Errorf("must specify one of --from-arch, --from-current-arch, --from-file, --from-memory")
	}

	return params, nil
}
