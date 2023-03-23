package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LouisBrunner/mem-viz/pkg/commons"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type Args struct {
	FromJSONFile string
	FromJSONText string
	OutputFormat string
	OutputFile   string
	LoggingLevel logrus.Level
}

var fromSources = []string{
	"from-json-file",
	"from-json-text",
}

var FromCommonSourcesHelp = strings.Join(commons.MapSlice(fromSources, strconv.Quote), ", ")

const (
	OutputFormatGraphviz = "graphviz"
	OutputFormatLaTeX    = "latex"
	OutputFormatMarkdown = "markdown"
	OutputFormatText     = "text"
	OutputFormatASCII    = "ascii"
	OutputFormatJSON     = "json"
)

var outputFormats = []string{
	OutputFormatGraphviz,
	OutputFormatLaTeX,
	OutputFormatMarkdown,
	OutputFormatText,
	OutputFormatASCII,
	OutputFormatJSON,
}

var OutputFormatsHelp = strings.Join(commons.MapSlice(outputFormats, strconv.Quote), ", ")

func GetDefaultArgs() Args {
	return Args{
		OutputFormat: OutputFormatText,
		LoggingLevel: logrus.ErrorLevel,
	}
}

func ParseArgs[T any](params *Args, userParams *T, addMore func(params *T), addFrom func(params T) ([]bool, []string)) error {
	help := false
	loggingLevelStr := ""

	pflag.StringVar(&params.FromJSONFile, "from-json", "", "use the JSON output from a previous run, e.g. `./blocks.json` or `-` for stdin")
	pflag.StringVar(&params.FromJSONText, "from-json-text", "", fmt.Sprintf("use the JSON output from a previous run, e.g. `%s`", `{"Name": "foo"}`))
	pflag.StringVar(&params.OutputFormat, "output", params.OutputFormat, fmt.Sprintf("output format, one of: %s", OutputFormatsHelp))
	pflag.StringVarP(&params.OutputFile, "output-file", "o", "", "output file, e.g. `./blocks.dot`, defaults to stdout")
	pflag.StringVar(&loggingLevelStr, "logging-level", params.LoggingLevel.String(), fmt.Sprintf("logrus log level for internal debugging, e.g. %q", logrus.DebugLevel.String()))
	if addMore != nil {
		addMore(userParams)
	}
	pflag.BoolVarP(&help, "help", "h", false, "show this help message and exit")
	pflag.Parse()

	if help {
		pflag.Usage()
		return pflag.ErrHelp
	}

	if pflag.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(pflag.Args(), ", "))
	}

	// Check log level
	logLevel, err := logrus.ParseLevel(loggingLevelStr)
	if err != nil {
		return err
	}
	params.LoggingLevel = logLevel

	// Check output format
	if params.OutputFormat == "" {
		return fmt.Errorf("must specify an output format")
	}
	if !slices.Contains(outputFormats, params.OutputFormat) {
		return fmt.Errorf("invalid output format: %q, must be one of %s", params.OutputFormat, OutputFormatsHelp)
	}

	// Check that only one of the --from-* flags is set
	var checks []bool
	var names []string
	if addFrom != nil {
		checks, names = addFrom(*userParams)
	}
	fromFlagsCount := map[bool]int{}
	fromFlagsEach := append([]bool{
		params.FromJSONFile != "",
		params.FromJSONText != "",
	}, checks...)
	for _, flag := range fromFlagsEach {
		fromFlagsCount[flag] += 1
	}

	var fromSourcesHelp = strings.Join(commons.MapSlice(append(fromSources, names...), strconv.Quote), ", ")
	if fromFlagsCount[true] > 1 {
		return fmt.Errorf("cannot specify more than one of %s", fromSourcesHelp)
	}
	if fromFlagsCount[true] == 0 {
		return fmt.Errorf("must specify one of %s", fromSourcesHelp)
	}

	return nil
}
