package main

import (
	"fmt"
	"strings"

	"github.com/LouisBrunner/dsc-viz/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type args struct {
	fromArch        string
	fromCurrentArch bool
	fromFile        string
	fromMemory      bool
	outputFormat    string
	loggingLevel    logrus.Level
}

const (
	outputFormatGraphviz = "graphviz"
	outputFormatLaTeX    = "latex"
	outputFormatMarkdown = "markdown"
	outputFormatText     = "text"
	outputFormatASCII    = "ascii"
	outputFormatJSON     = "json"
)

var outputFormats = []string{
	outputFormatGraphviz,
	outputFormatLaTeX,
	outputFormatMarkdown,
	outputFormatText,
	outputFormatASCII,
	outputFormatJSON,
}

var outputFormatsHelp = strings.Join(utils.MapSlice(outputFormats, func(s string) string {
	return fmt.Sprintf("%q", s)
}), ", ")

func getArgs() (*args, error) {
	params := &args{
		outputFormat: outputFormatText,
		loggingLevel: logrus.ErrorLevel,
	}

	pflag.StringVar(&params.fromArch, "from-arch", "", "scan your system to find the cache for the given architecture (e.g. `arm64`)")
	pflag.BoolVar(&params.fromCurrentArch, "from-current-arch", false, "scan your system to find the cache for the current architecture")
	pflag.StringVar(&params.fromFile, "from-file", "", "file to fetch from, e.g. `./dyld_shared_cache_arm64`")
	pflag.BoolVar(&params.fromMemory, "from-memory", false, "fetch from memory")
	pflag.StringVar(&params.outputFormat, "output", params.outputFormat, fmt.Sprintf("output format, one of: %s", outputFormatsHelp))
	loggingLevelStr := ""
	pflag.StringVar(&loggingLevelStr, "logging-level", params.loggingLevel.String(), fmt.Sprintf("logrus log level for internal debugging, e.g. %q", logrus.DebugLevel.String()))
	pflag.Parse()

	if pflag.NArg() > 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(pflag.Args(), ", "))
	}

	// Check log level
	logLevel, err := logrus.ParseLevel(loggingLevelStr)
	if err != nil {
		return nil, err
	}
	params.loggingLevel = logLevel

	// Check output format
	if params.outputFormat == "" {
		return nil, fmt.Errorf("must specify an output format")
	}
	if !slices.Contains(outputFormats, params.outputFormat) {
		return nil, fmt.Errorf("invalid output format: %q, must be one of %s", params.outputFormat, outputFormatsHelp)
	}

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
